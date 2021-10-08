// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	"os"
	"strings"
	"time"

	"fybrik.io/fybrik/manager/controllers"
	"fybrik.io/fybrik/pkg/environment"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	connectors "fybrik.io/fybrik/pkg/connectors/clients"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app/modules"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/storage"
	"fybrik.io/fybrik/pkg/vault"
)

// FybrikApplicationReconciler reconciles a FybrikApplication object
type FybrikApplicationReconciler struct {
	client.Client
	Name              string
	Log               logr.Logger
	Scheme            *runtime.Scheme
	PolicyManager     connectors.PolicyManager
	DataCatalog       connectors.DataCatalog
	ResourceInterface ContextInterface
	ClusterManager    multicluster.ClusterLister
	Provision         storage.ProvisionInterface
}

// Reconcile reconciles FybrikApplication CRD
// It receives FybrikApplication CRD and selects the appropriate modules that will run
// The outcome is a Plotter containing multiple Blueprints that run on different clusters
func (r *FybrikApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("fybrikapplication", req.NamespacedName)
	// obtain FybrikApplication resource
	applicationContext := &api.FybrikApplication{}
	if err := r.Get(ctx, req.NamespacedName, applicationContext); err != nil {
		log.V(0).Info("The reconciled object was not found")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.reconcileFinalizers(applicationContext); err != nil {
		log.V(0).Info("Could not reconcile finalizers " + err.Error())
		return ctrl.Result{}, err
	}

	// If the object has a scheduled deletion time, update status and return
	if !applicationContext.DeletionTimestamp.IsZero() {
		// The object is being deleted
		return ctrl.Result{}, nil
	}

	observedStatus := applicationContext.Status.DeepCopy()
	appVersion := applicationContext.GetGeneration()

	// check if webhooks are enabled and application has been validated before or if validated application is outdated
	if os.Getenv("ENABLE_WEBHOOKS") != "true" && (string(applicationContext.Status.ValidApplication) == "" || observedStatus.ValidatedGeneration != appVersion) {
		// do validation on applicationContext
		err := applicationContext.ValidateFybrikApplication("/tmp/taxonomy/fybrik_application.json")
		log.V(0).Info("Reconciler validating Fybrik application")
		applicationContext.Status.ValidatedGeneration = appVersion
		// if validation fails
		if err != nil {
			// set error message
			log.V(0).Info("Fybrik application validation failed " + err.Error())
			applicationContext.Status.ErrorMessage = err.Error()
			applicationContext.Status.ValidApplication = v1.ConditionFalse
			if err := r.Client.Status().Update(ctx, applicationContext); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		applicationContext.Status.ValidApplication = v1.ConditionTrue
	}
	if applicationContext.Status.ValidApplication == v1.ConditionFalse {
		return ctrl.Result{}, nil
	}

	// check if reconcile is required
	// reconcile is required if the spec has been changed, or the previous reconcile has failed to allocate a Plotter resource
	generationComplete := r.ResourceInterface.ResourceExists(observedStatus.Generated) && (observedStatus.Generated.AppVersion == appVersion)
	if (!generationComplete) || (observedStatus.ObservedGeneration != appVersion) {
		if result, err := r.reconcile(applicationContext); err != nil {
			// another attempt will be done
			// users should be informed in case of errors
			if !equality.Semantic.DeepEqual(&applicationContext.Status, observedStatus) {
				// ignore an update error, a new reconcile will be made in any case
				_ = r.Client.Status().Update(ctx, applicationContext)
			}
			return result, err
		}
		applicationContext.Status.ObservedGeneration = appVersion
	} else {
		resourceStatus, err := r.ResourceInterface.GetResourceStatus(applicationContext.Status.Generated)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err = r.checkReadiness(applicationContext, resourceStatus); err != nil {
			return ctrl.Result{}, err
		}
	}
	applicationContext.Status.Ready = isReady(applicationContext)

	// Update CRD status in case of change (other than deletion, which was handled separately)
	if !equality.Semantic.DeepEqual(&applicationContext.Status, observedStatus) && applicationContext.DeletionTimestamp.IsZero() {
		log.V(0).Info("Reconcile: Updating status for desired generation " + fmt.Sprint(applicationContext.GetGeneration()))
		if err := r.Client.Status().Update(ctx, applicationContext); err != nil {
			return ctrl.Result{}, err
		}
	}
	errorMsg := getErrorMessages(applicationContext)
	if errorMsg != "" {
		log.Info("Reconciled with errors: " + errorMsg)
	}

	// trigger a new reconcile if required (the fybrikapplication is not ready)
	if !isReady(applicationContext) {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

func getBucketResourceRef(name string) *types.NamespacedName {
	return &types.NamespacedName{Name: name, Namespace: utils.GetSystemNamespace()}
}

func (r *FybrikApplicationReconciler) checkReadiness(applicationContext *api.FybrikApplication, status api.ObservedState) error {
	if applicationContext.Status.AssetStates == nil {
		initStatus(applicationContext)
	}

	// TODO(shlomitk1): receive status per asset and update accordingly
	// Temporary fix: all assets that are not in Deny state are updated based on the received status
	for _, dataCtx := range applicationContext.Spec.Data {
		assetID := dataCtx.DataSetID
		if applicationContext.Status.AssetStates[assetID].Conditions[api.DenyConditionIndex].Status == v1.ConditionTrue {
			// should not appear in the plotter status
			continue
		}
		if status.Error != "" {
			setErrorCondition(applicationContext, assetID, status.Error)
			continue
		}
		if !status.Ready {
			continue
		}

		// register assets if necessary if the ready state has been received
		if dataCtx.Requirements.Copy.Catalog.CatalogID != "" {
			if applicationContext.Status.AssetStates[assetID].CatalogedAsset != "" {
				// the asset has been already cataloged
				continue
			}
			// mark the bucket as persistent and register the asset
			provisionedBucketRef, found := applicationContext.Status.ProvisionedStorage[assetID]
			if !found {
				message := "No copy has been created for the asset " + assetID + " required to be registered"
				r.Log.V(0).Info(message)
				setErrorCondition(applicationContext, assetID, message)
				continue
			}
			if err := r.Provision.SetPersistent(getBucketResourceRef(provisionedBucketRef.DatasetRef), true); err != nil {
				setErrorCondition(applicationContext, assetID, err.Error())
				continue
			}
			// register the asset: experimental feature
			if newAssetID, err := r.RegisterAsset(dataCtx.Requirements.Copy.Catalog.CatalogID, &provisionedBucketRef, applicationContext); err == nil {
				state := applicationContext.Status.AssetStates[assetID]
				state.CatalogedAsset = newAssetID
				applicationContext.Status.AssetStates[assetID] = state
			} else {
				// log an error and make a new attempt to register the asset
				setErrorCondition(applicationContext, assetID, err.Error())
				continue
			}
		}
		setReadyCondition(applicationContext, assetID)
	}
	return nil
}

// reconcileFinalizers reconciles finalizers for FybrikApplication
func (r *FybrikApplicationReconciler) reconcileFinalizers(applicationContext *api.FybrikApplication) error {
	// finalizer
	finalizerName := r.Name + ".finalizer"
	hasFinalizer := ctrlutil.ContainsFinalizer(applicationContext, finalizerName)

	// If the object has a scheduled deletion time, delete it and all resources it has created
	if !applicationContext.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if hasFinalizer { // Finalizer was created when the object was created
			// the finalizer is present - delete the allocated resources
			if err := r.deleteExternalResources(applicationContext); err != nil {
				return err
			}

			// remove the finalizer from the list and update it, because it needs to be deleted together with the object
			ctrlutil.RemoveFinalizer(applicationContext, finalizerName)

			if err := r.Update(context.Background(), applicationContext); err != nil {
				return err
			}
		}
		return nil
	}
	// Make sure this CRD instance has a finalizer
	if !hasFinalizer {
		ctrlutil.AddFinalizer(applicationContext, finalizerName)
		if err := r.Update(context.Background(), applicationContext); err != nil {
			return err
		}
	}
	return nil
}

func (r *FybrikApplicationReconciler) deleteExternalResources(applicationContext *api.FybrikApplication) error {
	// clear provisioned storage
	// References to buckets (Dataset resources) are deleted. Buckets that are persistent will not be removed upon Dataset deletion.
	var deletedKeys []string
	var errMsgs []string
	for datasetID, datasetDetails := range applicationContext.Status.ProvisionedStorage {
		if err := r.Provision.DeleteDataset(getBucketResourceRef(datasetDetails.DatasetRef)); err != nil {
			errMsgs = append(errMsgs, err.Error())
		} else {
			deletedKeys = append(deletedKeys, datasetID)
		}
	}
	for _, datasetID := range deletedKeys {
		delete(applicationContext.Status.ProvisionedStorage, datasetID)
	}
	if len(errMsgs) != 0 {
		return errors.New(strings.Join(errMsgs, ";"))
	}
	// delete the generated resource
	if applicationContext.Status.Generated == nil {
		return nil
	}

	r.Log.V(0).Info("Reconcile: FybrikApplication is deleting the generated " + applicationContext.Status.Generated.Kind)
	if err := r.ResourceInterface.DeleteResource(applicationContext.Status.Generated); err != nil {
		return err
	}
	applicationContext.Status.Generated = nil
	return nil
}

// setReadModulesEndpoints populates the ReadEndpointsMap map in the status of the fybrikapplication
func setReadModulesEndpoints(applicationContext *api.FybrikApplication, flows []api.Flow) {
	readEndpointMap := make(map[string]api.EndpointSpec)
	for _, flow := range flows {
		if flow.FlowType == api.ReadFlow {
			for _, subflow := range flow.SubFlows {
				if subflow.FlowType == api.ReadFlow {
					for _, sequentialSteps := range subflow.Steps {
						// Check the last step in the sequential flow that is for read (this will expose the reading api)
						lastStep := sequentialSteps[len(sequentialSteps)-1]
						if lastStep.Parameters.API != nil {
							readEndpointMap[flow.AssetID] = lastStep.Parameters.API.Endpoint
						}
					}
				}
			}
		}
	}
	// populate endpoints in application status
	for _, asset := range applicationContext.Spec.Data {
		id := utils.CreateDataSetIdentifier(asset.DataSetID)
		state := applicationContext.Status.AssetStates[asset.DataSetID]
		state.Endpoint = readEndpointMap[id]
		applicationContext.Status.AssetStates[asset.DataSetID] = state
	}
}

// reconcile receives either FybrikApplication CRD
// or a status update from the generated resource
func (r *FybrikApplicationReconciler) reconcile(applicationContext *api.FybrikApplication) (ctrl.Result, error) {
	utils.PrintStructure(applicationContext.Spec, r.Log, "FybrikApplication")
	// Data User created or updated the FybrikApplication

	// clear status
	initStatus(applicationContext)
	if applicationContext.Status.ProvisionedStorage == nil {
		applicationContext.Status.ProvisionedStorage = make(map[string]api.DatasetDetails)
	}

	if len(applicationContext.Spec.Data) == 0 {
		if err := r.deleteExternalResources(applicationContext); err != nil {
			return ctrl.Result{}, err
		}
		r.Log.V(0).Info("no blueprint will be generated since no datasets are specified")
		return ctrl.Result{}, nil
	}

	clusters, err := r.ClusterManager.GetClusters()
	if err != nil {
		return ctrl.Result{}, err
	}
	// create a list of requirements for creating a data flow (actions, interface to app, data format) per a single data set
	var requirements []modules.DataInfo
	for _, dataset := range applicationContext.Spec.Data {
		req := modules.DataInfo{
			Context: dataset.DeepCopy(),
		}
		if err := r.constructDataInfo(&req, applicationContext, clusters); err != nil {
			AnalyzeError(applicationContext, req.Context.DataSetID, err)
			continue
		}
		requirements = append(requirements, req)
	}
	// check if can proceed
	if len(requirements) == 0 {
		return ctrl.Result{}, nil
	}

	// create a module manager that will select modules to be orchestrated based on user requirements and module capabilities
	moduleMap, err := r.GetAllModules()
	if err != nil {
		return ctrl.Result{}, err
	}
	objectKey := client.ObjectKeyFromObject(applicationContext)
	moduleManager := &ModuleManager{
		Client:             r.Client,
		Log:                r.Log,
		Modules:            moduleMap,
		Clusters:           clusters,
		Owner:              objectKey,
		PolicyManager:      r.PolicyManager,
		Provision:          r.Provision,
		ProvisionedStorage: make(map[string]NewAssetInfo),
	}

	plotterSpec := &api.PlotterSpec{
		Selector:  applicationContext.Spec.Selector,
		Assets:    map[string]api.AssetDetails{},
		Flows:     []api.Flow{},
		Templates: map[string]api.Template{},
	}

	for _, item := range requirements {
		// TODO support different flows than read by specifying it in the application
		flowType := api.ReadFlow

		err := moduleManager.AddFlowInfoForAsset(item, applicationContext, plotterSpec, flowType)
		if err != nil {
			AnalyzeError(applicationContext, item.Context.DataSetID, err)
			continue
		}
	}
	// check if can proceed
	if getErrorMessages(applicationContext) != "" {
		return ctrl.Result{}, nil
	}

	// update allocated storage in the status
	// clean irrelevant buckets
	for datasetID, details := range applicationContext.Status.ProvisionedStorage {
		if _, found := moduleManager.ProvisionedStorage[datasetID]; !found {
			_ = r.Provision.DeleteDataset(getBucketResourceRef(details.DatasetRef))
			delete(applicationContext.Status.ProvisionedStorage, datasetID)
		}
	}
	// add or update new buckets
	for datasetID, info := range moduleManager.ProvisionedStorage {
		raw := serde.NewArbitrary(info.Details)
		applicationContext.Status.ProvisionedStorage[datasetID] = api.DatasetDetails{
			DatasetRef: info.Storage.Name,
			SecretRef:  info.Storage.SecretRef.Name,
			Details:    *raw,
		}
	}
	ready := true
	var allocErr error
	// check that the buckets have been created successfully using Dataset status
	for id, details := range applicationContext.Status.ProvisionedStorage {
		res, err := r.Provision.GetDatasetStatus(getBucketResourceRef(details.DatasetRef))
		if err != nil {
			ready = false
			break
		}
		if !res.Provisioned {
			ready = false
			r.Log.V(0).Info("No bucket has been provisioned for " + id)
			// TODO(shlomitk1): analyze the error
			if res.ErrorMsg != "" {
				allocErr = errors.New(res.ErrorMsg)
			}
			break
		}
	}
	if !ready {
		return ctrl.Result{RequeueAfter: 2 * time.Second}, allocErr
	}

	setReadModulesEndpoints(applicationContext, plotterSpec.Flows)
	ownerRef := &api.ResourceReference{Name: applicationContext.Name, Namespace: applicationContext.Namespace, AppVersion: applicationContext.GetGeneration()}

	resourceRef := r.ResourceInterface.CreateResourceReference(ownerRef)
	if err := r.ResourceInterface.CreateOrUpdateResource(ownerRef, resourceRef, plotterSpec, applicationContext.Labels); err != nil {
		r.Log.V(0).Info("Error creating " + resourceRef.Kind + " : " + err.Error())
		if err.Error() == api.InvalidClusterConfiguration {
			applicationContext.Status.ErrorMessage = err.Error()
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	applicationContext.Status.Generated = resourceRef
	r.Log.V(0).Info("Created " + resourceRef.Kind + " successfully!")
	return ctrl.Result{}, nil
}

func (r *FybrikApplicationReconciler) constructDataInfo(req *modules.DataInfo, input *api.FybrikApplication, clusters []multicluster.Cluster) error {
	var err error

	// Call the DataCatalog service to get info about the dataset
	var response *pb.CatalogDatasetInfo
	var credentialPath string
	if input.Spec.SecretRef != "" {
		credentialPath = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if response, err = r.DataCatalog.GetDatasetInfo(ctx, &pb.CatalogDatasetRequest{
		CredentialPath: credentialPath,
		DatasetId:      req.Context.DataSetID,
	}); err != nil {
		return err
	}

	details := response.GetDetails()
	dataDetails, err := modules.CatalogDatasetToDataDetails(response)
	if err != nil {
		return err
	}
	req.DataDetails = dataDetails
	req.VaultSecretPath = ""
	if details.CredentialsInfo != nil {
		req.VaultSecretPath = details.CredentialsInfo.VaultSecretPath
	}

	return nil
}

// NewFybrikApplicationReconciler creates a new reconciler for FybrikApplications
func NewFybrikApplicationReconciler(mgr ctrl.Manager, name string,
	policyManager connectors.PolicyManager, catalog connectors.DataCatalog, cm multicluster.ClusterLister, provision storage.ProvisionInterface) *FybrikApplicationReconciler {
	return &FybrikApplicationReconciler{
		Client:            mgr.GetClient(),
		Name:              name,
		Log:               ctrl.Log.WithName("controllers").WithName(name),
		Scheme:            mgr.GetScheme(),
		PolicyManager:     policyManager,
		ResourceInterface: NewPlotterInterface(mgr.GetClient()),
		ClusterManager:    cm,
		Provision:         provision,
		DataCatalog:       catalog,
	}
}

// SetupWithManager registers FybrikApplication controller
func (r *FybrikApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mapFn := func(a client.Object) []reconcile.Request {
		labels := a.GetLabels()
		if labels == nil {
			return []reconcile.Request{}
		}
		namespace, foundNamespace := labels[api.ApplicationNamespaceLabel]
		name, foundName := labels[api.ApplicationNameLabel]
		if !foundNamespace || !foundName {
			return []reconcile.Request{}
		}
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}},
		}
	}

	numReconciles := environment.GetEnvAsInt(controllers.ApplicationConcurrentReconcilesConfiguration, controllers.DefaultApplicationConcurrentReconciles)

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: numReconciles}).
		For(&api.FybrikApplication{}).
		Watches(&source.Kind{
			Type: &api.Plotter{},
		}, handler.EnqueueRequestsFromMapFunc(mapFn)).Complete(r)
}

// AnalyzeError analyzes whether the given error is fatal, or a retrial attempt can be made.
// Reasons for retrial can be either communication problems with external services, or kubernetes problems to perform some action on a resource.
// A retrial is achieved by returning an error to the reconcile method
func AnalyzeError(application *api.FybrikApplication, assetID string, err error) {
	if err == nil {
		return
	}
	switch err.Error() {
	case api.InvalidAssetID, api.ReadAccessDenied, api.CopyNotAllowed, api.WriteNotAllowed, api.InvalidAssetDataStore:
		setDenyCondition(application, assetID, err.Error())
	default:
		setErrorCondition(application, assetID, err.Error())
	}
}

func ownerLabels(id types.NamespacedName) map[string]string {
	return map[string]string{
		api.ApplicationNamespaceLabel: id.Namespace,
		api.ApplicationNameLabel:      id.Name,
	}
}

// GetAllModules returns all CRDs of the kind FybrikModule mapped by their name
func (r *FybrikApplicationReconciler) GetAllModules() (map[string]*api.FybrikModule, error) {
	ctx := context.Background()

	moduleMap := make(map[string]*api.FybrikModule)
	var moduleList api.FybrikModuleList
	if err := r.List(ctx, &moduleList, client.InNamespace(utils.GetSystemNamespace())); err != nil {
		r.Log.V(0).Info("Error while listing modules: " + err.Error())
		return moduleMap, err
	}
	r.Log.Info("Listing all modules")
	for _, module := range moduleList.Items {
		r.Log.Info(module.GetName())
		moduleMap[module.Name] = module.DeepCopy()
	}
	return moduleMap, nil
}
