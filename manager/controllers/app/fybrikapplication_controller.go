// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	api "fybrik.io/fybrik/manager/apis/app/v1"
	"fybrik.io/fybrik/manager/controllers"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	dcclient "fybrik.io/fybrik/pkg/connectors/datacatalog/clients"
	pmclient "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/policymanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/storage"
	"fybrik.io/fybrik/pkg/taxonomy/validate"
	"fybrik.io/fybrik/pkg/vault"
)

// FybrikApplicationReconciler reconciles a FybrikApplication object
type FybrikApplicationReconciler struct {
	client.Client
	Name              string
	Log               zerolog.Logger
	Scheme            *runtime.Scheme
	PolicyManager     pmclient.PolicyManager
	DataCatalog       dcclient.DataCatalog
	ResourceInterface ContextInterface
	ClusterManager    multicluster.ClusterLister
	Provision         storage.ProvisionInterface
	ConfigEvaluator   adminconfig.EvaluatorInterface
	Infrastructure    *infrastructure.AttributeManager
}

type ApplicationContext struct {
	Log         *zerolog.Logger
	Application *api.FybrikApplication
	UUID        string
}

var ApplicationTaxonomy = environment.GetDataDir() + "/taxonomy/fybrik_application.json"
var DataCatalogTaxonomy = environment.GetDataDir() + "/taxonomy/datacatalog.json#/definitions/GetAssetResponse"

const (
	FybrikApplicationKind = "FybrikApplication"
	PlotterUpdatePrefix   = "plotter_"
	Interval              = 10
)

// ErrorMessages that are reported to the user
const (
	InvalidAssetID              string = "the asset does not exist"
	ReadAccessDenied            string = "governance policies forbid access to the data"
	CopyNotAllowed              string = "copy of the data is required but can not be done according to the governance policies"
	WriteNotAllowed             string = "governance policies forbid writing of the data"
	StorageAccountUndefined     string = "no storage account has been defined"
	ModuleNotFound              string = "no module has been registered"
	InsufficientStorage         string = "no bucket was provisioned for implicit copy"
	InvalidClusterConfiguration string = "cluster configuration does not support the requirements"
	InvalidAssetDataStore       string = "the asset data store is not supported"
)

// Reconcile reconciles FybrikApplication CRD
// It receives FybrikApplication CRD and selects the appropriate modules that will run
// The outcome is a Plotter containing multiple Blueprints that run on different clusters
//nolint:gocyclo
func (r *FybrikApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	sublog := r.Log.With().Str(FybrikApplicationKind, req.NamespacedName.String()).Logger()

	sublog.Trace().Msg("*** FybrikApplication Reconcile ***")
	// obtain FybrikApplication resource
	// events coming from plotter updates have a special prefix prepended to the name of fybrik application
	plotterUpdate := false
	nsName := req.NamespacedName
	if strings.HasPrefix(nsName.Name, PlotterUpdatePrefix) {
		// reconcile results from plotter changes
		plotterUpdate = true
		nsName.Name = nsName.Name[len(PlotterUpdatePrefix):]
	}
	application := &api.FybrikApplication{}
	if err := r.Get(ctx, nsName, application); err != nil {
		sublog.Warn().Msg("The reconciled object was not found")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	uuid := utils.GetFybrikApplicationUUID(application)
	log := sublog.With().Str(utils.FybrikAppUUID, uuid).Logger()

	// Log the fybrikapplication
	logging.LogStructure(FybrikApplicationKind, application, &log, zerolog.TraceLevel, true, true)
	applicationContext := ApplicationContext{Log: &log, Application: application, UUID: uuid}
	if plotterUpdate && (application.Status.Generated == nil || application.Status.Generated.AppVersion != application.GetGeneration()) {
		// plotter update has been received but it does not match the fybrik application status
		// this can happen if the plotter has just been created, and the application status was not updated by the server
		// ignore and wait for the next plotter update
		log.Debug().Msg("Ignoring plotter update")
		return ctrl.Result{}, nil
	}

	// If the object has a scheduled deletion time, delete it and all resources it has created
	if !applicationContext.Application.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.removeFinalizers(ctx, applicationContext)
	}

	observedStatus := application.Status.DeepCopy()
	appVersion := application.GetGeneration()

	// validate fybrik application in case of the create/update resource event
	if err := r.validateApp(ctx, applicationContext); err != nil {
		return ctrl.Result{}, err
	}
	if application.Status.ValidApplication == v1.ConditionFalse {
		return ctrl.Result{}, nil
	}

	// no datasets are specified - remove finalizers and old resources
	if len(applicationContext.Application.Spec.Data) == 0 {
		if err := r.removeFinalizers(ctx, applicationContext); err != nil {
			return ctrl.Result{}, err
		}
		applicationContext.Log.Info().Msg("No plotter will be generated since no datasets are specified")
		return ctrl.Result{}, nil
	}

	// check if reconcile is required
	// reconcile is required if the spec has been changed, or the previous reconcile has failed to allocate a Plotter resource
	generationComplete := observedStatus.Generated != nil && (observedStatus.Generated.AppVersion == appVersion)
	if plotterUpdate {
		// check plotter status and update the application status accordingly
		resourceStatus, err := r.ResourceInterface.GetResourceStatus(application.Status.Generated)
		if err != nil {
			return ctrl.Result{}, err
		}
		r.checkReadiness(applicationContext, resourceStatus)
	} else if (observedStatus.ObservedGeneration != appVersion) || !generationComplete {
		// spec has been changed, or there was a failure to allocate a plotter
		if result, err := r.reconcile(applicationContext); err != nil || result.Requeue || (result.RequeueAfter > 0) {
			// another attempt will be done
			// users should be informed in case of errors
			// ignore an update error, a new reconcile will be made in any case
			_ = utils.UpdateStatus(ctx, r.Client, application, observedStatus)
			return result, err
		}
		application.Status.ObservedGeneration = appVersion
	}
	application.Status.Ready = isReady(application)
	log.Trace().Str(logging.ACTION, logging.UPDATE).Msg("Updating status for desired generation " + fmt.Sprint(application.GetGeneration()))
	if err := utils.UpdateStatus(ctx, r.Client, application, observedStatus); err != nil {
		return ctrl.Result{}, err
	}
	// add finalizers if some resources have been allocated (plotter, datasets)
	if application.Status.Generated != nil || (len(application.Status.ProvisionedStorage) > 0) {
		if err := r.addFinalizers(ctx, applicationContext); err != nil {
			return ctrl.Result{}, err
		}
	}
	if errorMsg := getErrorMessages(application); errorMsg != "" {
		log.Warn().Str(logging.ACTION, logging.UPDATE).Msg("Reconcile failed with errors")
		// trigger a new reconcile
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func getBucketResourceRef(name string) *types.NamespacedName {
	return &types.NamespacedName{Name: name, Namespace: environment.GetSystemNamespace()}
}

func (r *FybrikApplicationReconciler) checkReadiness(applicationContext ApplicationContext, status api.ObservedState) {
	if applicationContext.Application.Status.AssetStates == nil {
		initStatus(applicationContext.Application)
	}

	// TODO(shlomitk1): receive status per asset and update accordingly
	// Temporary fix: all assets that are not in Deny state are updated based on the received status
	for _, dataCtx := range applicationContext.Application.Spec.Data {
		assetID := dataCtx.DataSetID
		if applicationContext.Application.Status.AssetStates[assetID].Conditions[DenyConditionIndex].Status == v1.ConditionTrue {
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
		if dataCtx.Requirements.FlowParams.Catalog != "" {
			if applicationContext.Application.Status.AssetStates[assetID].CatalogedAsset != "" {
				// the asset has been already cataloged
				continue
			}
			// mark the bucket as persistent and register the asset
			provisionedBucketRef, found := applicationContext.Application.Status.ProvisionedStorage[assetID]
			if !found {
				message := "No storage has been allocated for the asset " + assetID + " required to be registered"
				setErrorCondition(applicationContext, assetID, message)
				continue
			}
			if err := r.Provision.SetPersistent(getBucketResourceRef(provisionedBucketRef.DatasetRef), true); err != nil {
				setErrorCondition(applicationContext, assetID, err.Error())
				continue
			}
			reqResource := dataCtx.Requirements.FlowParams.ResourceMetadata
			if reqResource != nil {
				// we assume to have only the geography field set at this point
				// in the provisionedBucket ResourceMetadata
				geo := provisionedBucketRef.ResourceMetadata.Geography
				if reqResource.Geography != "" && geo != reqResource.Geography {
					// log conflict in Geography field
					applicationContext.Log.Warn().Msg("Geography field from application flow requirements " +
						"does not match provisioned bucket Geography and thus ignored")
				}
				provisionedBucketRef.ResourceMetadata = reqResource.DeepCopy()
				provisionedBucketRef.ResourceMetadata.Geography = geo
			}
			// register the asset
			if newAssetID, err := r.RegisterAsset(assetID, dataCtx.Requirements.FlowParams.Catalog,
				&provisionedBucketRef, applicationContext.Application); err == nil {
				state := applicationContext.Application.Status.AssetStates[assetID]
				state.CatalogedAsset = newAssetID
				applicationContext.Application.Status.AssetStates[assetID] = state
			} else {
				// log an error and make a new attempt to register the asset
				setErrorCondition(applicationContext, assetID, err.Error())
				continue
			}
		}
		setReadyCondition(applicationContext, assetID)
	}
}

func (r *FybrikApplicationReconciler) getFinalizerName() string {
	return r.Name + ".finalizer"
}

// removeFinalizers removes finalizers for FybrikApplication
func (r *FybrikApplicationReconciler) removeFinalizers(ctx context.Context, applicationContext ApplicationContext) error {
	// finalizer
	finalizerName := r.getFinalizerName()
	original := applicationContext.Application.DeepCopy()
	initStatus(applicationContext.Application)
	applicationContext.Application.Status.ObservedGeneration = applicationContext.Application.GetGeneration()
	if err := r.deleteExternalResources(applicationContext); err != nil {
		return err
	}
	if err := utils.UpdateStatus(ctx, r.Client, applicationContext.Application, &original.Status); err != nil {
		return err
	}
	if ctrlutil.ContainsFinalizer(applicationContext.Application, finalizerName) {
		// remove the finalizer from the list and update it, because it needs to be deleted together with the object
		ctrlutil.RemoveFinalizer(applicationContext.Application, finalizerName)
		if err := r.Patch(ctx, applicationContext.Application, client.MergeFrom(original)); err != nil {
			return err
		}
	}
	return nil
}

// addFinalizers adds finalizers for FybrikApplication
func (r *FybrikApplicationReconciler) addFinalizers(ctx context.Context, applicationContext ApplicationContext) error {
	// finalizer
	finalizerName := r.getFinalizerName()
	if !ctrlutil.ContainsFinalizer(applicationContext.Application, finalizerName) {
		original := applicationContext.Application.DeepCopy()
		ctrlutil.AddFinalizer(applicationContext.Application, finalizerName)
		// use Patch to preserve the generation version
		if err := r.Patch(ctx, applicationContext.Application, client.MergeFrom(original)); err != nil {
			return err
		}
	}
	return nil
}

func (r *FybrikApplicationReconciler) deleteExternalResources(applicationContext ApplicationContext) error {
	// clear provisioned storage
	// References to buckets (Dataset resources) are deleted. Buckets that are persistent will not be removed upon Dataset deletion.
	var deletedKeys []string
	var errMsgs []string
	for datasetID, datasetDetails := range applicationContext.Application.Status.ProvisionedStorage {
		if err := r.Provision.DeleteDataset(getBucketResourceRef(datasetDetails.DatasetRef)); err != nil {
			errMsgs = append(errMsgs, err.Error())
		} else {
			deletedKeys = append(deletedKeys, datasetID)
		}
	}
	for _, datasetID := range deletedKeys {
		delete(applicationContext.Application.Status.ProvisionedStorage, datasetID)
	}
	if len(errMsgs) != 0 {
		return errors.New(strings.Join(errMsgs, ";"))
	}
	// delete the generated resource
	if applicationContext.Application.Status.Generated == nil {
		return nil
	}

	applicationContext.Log.Trace().Str(logging.ACTION, logging.DELETE).
		Msgf("Reconcile: FybrikApplication is deleting the generated %s", applicationContext.Application.Status.Generated.Kind)
	if err := r.ResourceInterface.DeleteResource(applicationContext.Application.Status.Generated); err != nil {
		return err
	}
	applicationContext.Application.Status.Generated = nil
	return nil
}

// setVirtualEndpoints populates the endpoints in the status of the fybrikapplication
func setVirtualEndpoints(application *api.FybrikApplication, flows []api.Flow) {
	endpointMap := make(map[string]taxonomy.Connection)
	for _, flow := range flows {
		// sanity check
		if len(flow.SubFlows) == 0 {
			continue
		}
		subflow := flow.SubFlows[len(flow.SubFlows)-1]
		for _, sequentialSteps := range subflow.Steps {
			// Check the last step in the sequential flow (this will expose the api)
			lastStep := sequentialSteps[len(sequentialSteps)-1]
			if lastStep.Parameters.API != nil {
				endpointMap[flow.AssetID] = lastStep.Parameters.API.Connection
			}
		}
	}
	// populate endpoints in application status
	for _, asset := range application.Spec.Data {
		state := application.Status.AssetStates[asset.DataSetID]
		state.Endpoint = endpointMap[asset.DataSetID]
		application.Status.AssetStates[asset.DataSetID] = state
	}
}

// reconcile receives either FybrikApplication CRD
// or a status update from the generated resource
func (r *FybrikApplicationReconciler) reconcile(applicationContext ApplicationContext) (ctrl.Result, error) {
	// Log the request received - i.e. the fybrikapplication.spec
	applicationContext.Log.Trace().Msg("*** reconcile ***")

	// Data User created or updated the FybrikApplication

	// clear status
	initStatus(applicationContext.Application)
	if applicationContext.Application.Status.ProvisionedStorage == nil {
		applicationContext.Application.Status.ProvisionedStorage = make(map[string]api.DatasetDetails)
	}

	// create a list of requirements for creating a data flow (actions, interface to app, data format) per a single data set
	env, err := r.Environment()
	if err != nil {
		return ctrl.Result{}, err
	}
	// workload cluster is common for all datasets in the given application
	workloadCluster, err := r.GetWorkloadCluster(applicationContext, env)
	if err != nil {
		// fatal
		applicationContext.Log.Info().Err(err).Bool(logging.FORUSER, true).Bool(logging.AUDIT, true).
			Str(logging.ACTION, logging.CREATE).Msg("Could not determine in which cluster the workload runs")
		return ctrl.Result{}, err
	}
	var requirements []datapath.DataInfo
	for _, dataset := range applicationContext.Application.Spec.Data {
		req := datapath.DataInfo{
			Context:             dataset.DeepCopy(),
			DataDetails:         &datacatalog.GetAssetResponse{},
			StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		}
		if err = r.constructDataInfo(&req, applicationContext, workloadCluster, env); err != nil {
			AnalyzeError(applicationContext, req.Context.DataSetID, err)
			continue
		}
		requirements = append(requirements, req)
	}
	// check if can proceed
	if len(requirements) == 0 {
		return ctrl.Result{}, nil
	}

	provisionedStorage, plotterSpec, err := r.buildSolution(applicationContext, env, requirements)
	if err != nil {
		applicationContext.Log.Error().Err(err).Bool(logging.FORUSER, true).Bool(logging.AUDIT, true).Msg("Plotter construction failed")
	}
	// check if can proceed
	if err != nil || getErrorMessages(applicationContext.Application) != "" {
		return ctrl.Result{}, err
	}

	// clean irrelevant buckets and check that the provisioned storage is ready
	storageReady, allocationErr := r.updateProvisionedStorageStatus(applicationContext, provisionedStorage)
	if !storageReady {
		return ctrl.Result{RequeueAfter: 2 * time.Second}, allocationErr
	}

	setVirtualEndpoints(applicationContext.Application, plotterSpec.Flows)
	ownerRef := &api.ResourceReference{Name: applicationContext.Application.Name, Namespace: applicationContext.Application.Namespace,
		AppVersion: applicationContext.Application.GetGeneration()}

	resourceRef := r.ResourceInterface.CreateResourceReference(ownerRef)
	if err := r.ResourceInterface.CreateOrUpdateResource(ownerRef, resourceRef, plotterSpec,
		applicationContext.Application.Labels, applicationContext.UUID); err != nil {
		applicationContext.Log.Error().Err(err).Str(logging.ACTION, logging.CREATE).Msgf("Error creating %s", resourceRef.Kind)
		if err.Error() == InvalidClusterConfiguration {
			applicationContext.Application.Status.ErrorMessage = err.Error()
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	applicationContext.Application.Status.Generated = resourceRef
	applicationContext.Log.Trace().Str(logging.ACTION, logging.CREATE).Msgf("Created %s successfully!", resourceRef.Kind)
	return ctrl.Result{}, nil
}

func (r *FybrikApplicationReconciler) Environment() (*datapath.Environment, error) {
	// get deployed modules
	moduleMap, err := r.GetAllModules()
	if err != nil {
		r.Log.Error().Err(err).Msg("Error while listing modules")
		return nil, err
	}
	r.Log.Info().Msg("Listing modules")
	for m := range moduleMap {
		r.Log.Info().Msgf("Module: %s", m)
	}
	accounts, err := r.getStorageAccounts()
	if err != nil {
		r.Log.Error().Err(err).Msg("Error while listing storage accounts")
		return nil, err
	}
	// get available clusters
	clusters, err := r.ClusterManager.GetClusters()
	if err != nil {
		return nil, err
	}
	return &datapath.Environment{
		Modules:          moduleMap,
		Clusters:         clusters,
		StorageAccounts:  accounts,
		AttributeManager: r.Infrastructure,
	}, nil
}

// CreateDataRequest generates a new DataRequest object for a specific asset based on FybrikApplication and asset metadata
func CreateDataRequest(application *api.FybrikApplication, dataCtx *api.DataContext,
	assetMetadata *datacatalog.ResourceMetadata) adminconfig.DataRequest {
	var flow taxonomy.DataFlow

	// If a workload selector is provided but no flow, assume read - for backward compatibility
	if (application.Spec.Selector.WorkloadSelector.Size() > 0) && (dataCtx.Flow == "") {
		flow = taxonomy.ReadFlow
	} else {
		flow = dataCtx.Flow
	}
	return adminconfig.DataRequest{
		DatasetID: dataCtx.DataSetID,
		Interface: dataCtx.Requirements.Interface,
		Usage:     flow,
		Metadata:  assetMetadata,
	}
}

func (r *FybrikApplicationReconciler) ValidateAssetResponse(response *datacatalog.GetAssetResponse, taxonomyFile, datasetID string) error {
	var allErrs []*field.Error

	// Convert GetAssetRequest Go struct to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return err
	}
	r.Log.Info().Msg("responseJSON:" + string(responseJSON))

	// Validate Fybrik module against taxonomy
	allErrs, err = validate.TaxonomyCheck(responseJSON, taxonomyFile)
	if err != nil {
		return err
	}

	// Return any error
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.fybrik.io", Kind: "DataCatalog-AssetResponse"},
		datasetID, allErrs)
}

func (r *FybrikApplicationReconciler) constructDataInfo(req *datapath.DataInfo, appContext ApplicationContext,
	workloadCluster multicluster.Cluster, env *datapath.Environment) error {
	// Call the DataCatalog service to get info about the dataset
	input := appContext.Application
	log := appContext.Log.With().Str(logging.DATASETID, req.Context.DataSetID).Logger()
	var err error
	if !req.Context.Requirements.FlowParams.IsNewDataSet {
		var credentialPath string
		if input.Spec.SecretRef != "" {
			if !environment.IsVaultEnabled() {
				log.Error().Str("SecretRef", input.Spec.SecretRef).Msg("SecretRef defined [%s], but vault is disabled")
			} else {
				credentialPath = vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
			}
		}
		var response *datacatalog.GetAssetResponse
		request := datacatalog.GetAssetRequest{
			AssetID:       taxonomy.AssetID(req.Context.DataSetID),
			OperationType: datacatalog.READ}

		if response, err = r.DataCatalog.GetAssetInfo(&request, credentialPath); err != nil {
			log.Error().Err(err).Msg("failed to receive the catalog connector response")
			return err
		}

		err = r.ValidateAssetResponse(response, DataCatalogTaxonomy, req.Context.DataSetID)
		if err != nil {
			log.Error().Err(err).Msg("failed to validate the catalog connector response")
			return err
		}
		logging.LogStructure("Catalog connector response", response, &log, zerolog.DebugLevel, false, false)
		response.DeepCopyInto(req.DataDetails)
	} else if req.Context.Requirements.FlowParams.ResourceMetadata != nil {
		// Fill req.DataDetails with the metadata from the fybrikapplication
		req.DataDetails.ResourceMetadata = *req.Context.Requirements.FlowParams.ResourceMetadata
	}
	configEvaluatorInput := &adminconfig.EvaluatorInput{}
	configEvaluatorInput.Workload.UUID = utils.GetFybrikApplicationUUID(input)
	input.Spec.AppInfo.DeepCopyInto(&configEvaluatorInput.Workload.Properties)
	configEvaluatorInput.Workload.Cluster = workloadCluster
	configEvaluatorInput.Request = CreateDataRequest(input, req.Context, &req.DataDetails.ResourceMetadata)

	// Governance actions
	err = r.checkGovernanceActions(configEvaluatorInput, req, appContext, env)
	if err != nil {
		return err
	}
	configDecisions, err := r.ConfigEvaluator.Evaluate(configEvaluatorInput)
	if err != nil {
		appContext.Log.Error().Err(err).Msg("Error evaluating config policies")
		return err
	}
	logging.LogStructure("Config Policy Decisions", configDecisions, appContext.Log, zerolog.DebugLevel, false, false)
	req.WorkloadCluster = configEvaluatorInput.Workload.Cluster
	req.Configuration = configDecisions
	return nil
}

func (r *FybrikApplicationReconciler) checkGovernanceActions(configEvaluatorInput *adminconfig.EvaluatorInput,
	req *datapath.DataInfo, appContext ApplicationContext, env *datapath.Environment) error {
	var err error
	switch configEvaluatorInput.Request.Usage {
	case taxonomy.WriteFlow:
		if !req.Context.Requirements.FlowParams.IsNewDataSet {
			// update an existing dataset
			// query the policy manager whether the operation is allowed
			reqAction := policymanager.RequestAction{
				ActionType:         configEvaluatorInput.Request.Usage,
				Destination:        req.DataDetails.ResourceMetadata.Geography,
				ProcessingLocation: taxonomy.ProcessingLocation(configEvaluatorInput.Workload.Cluster.Metadata.Region),
			}
			req.Actions, err = LookupPolicyDecisions(req.Context.DataSetID, &req.DataDetails.ResourceMetadata,
				r.PolicyManager, appContext, &reqAction)
		}
	case taxonomy.ReadFlow, taxonomy.DeleteFlow:
		reqAction := policymanager.RequestAction{
			ActionType:         configEvaluatorInput.Request.Usage,
			Destination:        configEvaluatorInput.Workload.Cluster.Metadata.Region,
			ProcessingLocation: taxonomy.ProcessingLocation(configEvaluatorInput.Workload.Cluster.Metadata.Region),
		}
		req.Actions, err = LookupPolicyDecisions(req.Context.DataSetID, &req.DataDetails.ResourceMetadata,
			r.PolicyManager, appContext, &reqAction)
	}
	if err != nil {
		return err
	}
	var resMetadata *datacatalog.ResourceMetadata
	// query the policy manager whether WRITE operation is allowed
	if req.Context.Requirements.FlowParams.IsNewDataSet {
		if req.Context.Requirements.FlowParams.ResourceMetadata != nil {
			resMetadata = req.Context.Requirements.FlowParams.ResourceMetadata
		} else {
			resMetadata = &datacatalog.ResourceMetadata{
				Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{}}},
			}
		}
	} else {
		// Use the existsing resource metadata if the asset is not new
		resMetadata = &req.DataDetails.ResourceMetadata
	}
	for accountInd := range env.StorageAccounts {
		region := env.StorageAccounts[accountInd].Spec.Region
		reqAction := policymanager.RequestAction{
			ActionType:         taxonomy.WriteFlow,
			Destination:        string(region),
			ProcessingLocation: region,
		}

		actions, err := LookupPolicyDecisions(req.Context.DataSetID, resMetadata, r.PolicyManager, appContext, &reqAction)
		if err == nil {
			req.StorageRequirements[region] = actions
		} else if err.Error() != WriteNotAllowed {
			return err
		}
	}
	accountRequired := (req.Context.Requirements.FlowParams.IsNewDataSet && configEvaluatorInput.Request.Usage == taxonomy.WriteFlow) ||
		(configEvaluatorInput.Request.Usage == taxonomy.CopyFlow)
	// no account is defined, return an error for write and copy flows
	if len(env.StorageAccounts) == 0 && accountRequired {
		return errors.New(StorageAccountUndefined)
	}
	// write is denied to all accounts, return Deny for write and copy flows
	if len(req.StorageRequirements) == 0 && accountRequired {
		return errors.New(WriteNotAllowed)
	}
	return nil
}

// GetWorkloadCluster returns a workload cluster
// If no cluster has been specified for a workload, a local cluster is assumed.
func (r *FybrikApplicationReconciler) GetWorkloadCluster(appContext ApplicationContext,
	env *datapath.Environment) (multicluster.Cluster, error) {
	clusterName := appContext.Application.Spec.Selector.ClusterName
	if clusterName == "" {
		// if no workload selector is specified - it is not a read scenario, skip
		if appContext.Application.Spec.Selector.WorkloadSelector.Size() == 0 {
			return multicluster.Cluster{}, nil
		}
		// the workload runs in a local cluster
		appContext.Log.Warn().Err(errors.New("selector.clusterName field is not specified")).
			Str(logging.ACTION, logging.CREATE).Msg("No workload cluster indicated, so a local cluster is assumed")
		clusterName = environment.GetLocalClusterName()
	}
	// find the cluster by its name as it is specified in FybrikApplication workload selector
	for _, cluster := range env.Clusters {
		if cluster.Name == clusterName {
			return cluster, nil
		}
	}
	return multicluster.Cluster{}, errors.New("Cluster " + clusterName + " is not available")
}

// NewFybrikApplicationReconciler creates a new reconciler for FybrikApplications
func NewFybrikApplicationReconciler(mgr ctrl.Manager, name string,
	policyManager pmclient.PolicyManager, catalog dcclient.DataCatalog, cm multicluster.ClusterLister,
	provision storage.ProvisionInterface, evaluator adminconfig.EvaluatorInterface,
	attributeManager *infrastructure.AttributeManager) *FybrikApplicationReconciler {
	log := logging.LogInit(logging.CONTROLLER, name)
	return &FybrikApplicationReconciler{
		Client:            mgr.GetClient(),
		Name:              name,
		Log:               log,
		Scheme:            mgr.GetScheme(),
		PolicyManager:     policyManager,
		ResourceInterface: NewPlotterInterface(mgr.GetClient()),
		ClusterManager:    cm,
		Provision:         provision,
		DataCatalog:       catalog,
		ConfigEvaluator:   evaluator,
		Infrastructure:    attributeManager,
	}
}

// SetupWithManager registers FybrikApplication controller
func (r *FybrikApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mapFn := func(a client.Object) []reconcile.Request {
		labels := a.GetLabels()
		if labels == nil {
			return []reconcile.Request{}
		}
		if !a.GetDeletionTimestamp().IsZero() {
			// the owned resource is deleted - no updates should be sent
			return []reconcile.Request{}
		}
		namespace := utils.GetApplicationNamespaceFromLabels(labels)
		name := utils.GetApplicationNameFromLabels(labels)
		if namespace == "" || name == "" {
			return []reconcile.Request{}
		}
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Name:      PlotterUpdatePrefix + name,
				Namespace: namespace,
			}},
		}
	}

	numReconciles := environment.GetEnvAsInt(controllers.ApplicationConcurrentReconcilesConfiguration,
		controllers.DefaultApplicationConcurrentReconciles)

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: numReconciles}).
		For(&api.FybrikApplication{}).
		Watches(&source.Kind{
			Type: &api.Plotter{},
		}, handler.EnqueueRequestsFromMapFunc(mapFn)).Complete(r)
}

// AnalyzeError analyzes whether the given error is fatal, or a retrial attempt can be made.
// Reasons for retrial can be either communication problems with external services, or kubernetes
// problems to perform some action on a resource.
// A retrial is achieved by returning an error to the reconcile method
func AnalyzeError(appContext ApplicationContext, assetID string, err error) {
	if err == nil {
		return
	}
	switch err.Error() {
	case InvalidAssetID, ReadAccessDenied, CopyNotAllowed, WriteNotAllowed, InvalidAssetDataStore:
		setDenyCondition(appContext, assetID, err.Error())
	default:
		setErrorCondition(appContext, assetID, err.Error())
	}
}

func ownerLabels(id types.NamespacedName) map[string]string {
	return map[string]string{
		utils.ApplicationNamespaceLabel: id.Namespace,
		utils.ApplicationNameLabel:      id.Name,
	}
}

// GetAllModules returns all CRDs of the kind FybrikModule mapped by their name
func (r *FybrikApplicationReconciler) GetAllModules() (map[string]*api.FybrikModule, error) {
	ctx := context.Background()
	moduleMap := make(map[string]*api.FybrikModule)
	var moduleList api.FybrikModuleList
	if err := r.List(ctx, &moduleList, client.InNamespace(environment.GetSystemNamespace())); err != nil {
		return moduleMap, err
	}
	for ind := range moduleList.Items {
		moduleMap[moduleList.Items[ind].Name] = &moduleList.Items[ind]
	}
	return moduleMap, nil
}

// get all available storage accounts
func (r *FybrikApplicationReconciler) getStorageAccounts() ([]*api.FybrikStorageAccount, error) {
	var accountList api.FybrikStorageAccountList
	if err := r.List(context.Background(), &accountList, client.InNamespace(environment.GetSystemNamespace())); err != nil {
		return nil, err
	}
	accounts := []*api.FybrikStorageAccount{}
	for i := range accountList.Items {
		accounts = append(accounts, accountList.Items[i].DeepCopy())
	}
	return accounts, nil
}

func (r *FybrikApplicationReconciler) updateProvisionedStorageStatus(applicationContext ApplicationContext,
	provisionedStorage map[string]NewAssetInfo) (bool, error) {
	// update allocated storage in the status
	// clean irrelevant buckets
	for datasetID, details := range applicationContext.Application.Status.ProvisionedStorage {
		if _, found := provisionedStorage[datasetID]; !found {
			_ = r.Provision.DeleteDataset(getBucketResourceRef(details.DatasetRef))
			delete(applicationContext.Application.Status.ProvisionedStorage, datasetID)
		}
	}
	// add or update new buckets
	for datasetID, info := range provisionedStorage {
		details := &api.DataStore{}
		if info.Details != nil {
			details = info.Details.DeepCopy()
		}

		applicationContext.Application.Status.ProvisionedStorage[datasetID] = api.DatasetDetails{
			DatasetRef:       info.Storage.Name,
			SecretRef:        api.SecretRef{Name: info.Storage.SecretRef.Name, Namespace: info.Storage.SecretRef.Namespace},
			Details:          details,
			ResourceMetadata: &datacatalog.ResourceMetadata{Geography: info.Storage.Region},
		}
	}
	// check that the buckets have been created successfully using Dataset status
	for id, details := range applicationContext.Application.Status.ProvisionedStorage {
		res, err := r.Provision.GetDatasetStatus(getBucketResourceRef(details.DatasetRef))
		if err != nil {
			return false, nil
		}
		if !res.Provisioned {
			applicationContext.Log.Warn().Err(errors.New(res.ErrorMsg)).Str(logging.ACTION, logging.CREATE).
				Str(logging.DATASETID, id).Msg("No bucket has been provisioned")
			if res.ErrorMsg != "" {
				return false, errors.New(res.ErrorMsg)
			}
			return false, nil
		}
	}
	return true, nil
}

func (r *FybrikApplicationReconciler) buildSolution(applicationContext ApplicationContext, env *datapath.Environment,
	requirements []datapath.DataInfo) (map[string]NewAssetInfo, *api.PlotterSpec, error) {
	plotterGen := &PlotterGenerator{
		Client:             r.Client,
		Log:                applicationContext.Log,
		Owner:              client.ObjectKeyFromObject(applicationContext.Application),
		Provision:          r.Provision,
		ProvisionedStorage: make(map[string]NewAssetInfo),
	}

	plotterSpec := &api.PlotterSpec{
		Selector:         applicationContext.Application.Spec.Selector,
		AppInfo:          applicationContext.Application.Spec.AppInfo,
		Assets:           map[string]api.AssetDetails{},
		Flows:            []api.Flow{},
		ModulesNamespace: environment.GetDefaultModulesNamespace(),
		Templates:        map[string]api.Template{},
	}

	paths, err := solve(env, requirements, applicationContext.Log)
	if err != nil {
		applicationContext.Application.Status.ErrorMessage = err.Error()
		return plotterGen.ProvisionedStorage, plotterSpec, nil
	}
	if len(paths) != len(requirements) {
		return plotterGen.ProvisionedStorage, plotterSpec, errors.New("Wrong number of data paths")
	}

	for ind := range requirements {
		// If the flag IsNewDataSet is true then a new asset must be allocated
		if requirements[ind].Context.Requirements.FlowParams.IsNewDataSet {
			err = plotterGen.handleNewAsset(&requirements[ind], &paths[ind])
			if err != nil {
				setErrorCondition(applicationContext, requirements[ind].Context.DataSetID, err.Error())
				return plotterGen.ProvisionedStorage, plotterSpec, err
			}
		}
		err = plotterGen.AddFlowInfoForAsset(&requirements[ind], applicationContext.Application, &paths[ind], plotterSpec)
		if err != nil {
			setErrorCondition(applicationContext, requirements[ind].Context.DataSetID, err.Error())
			return plotterGen.ProvisionedStorage, plotterSpec, err
		}
	}
	return plotterGen.ProvisionedStorage, plotterSpec, nil
}

// validation of FybrikApplication
func (r *FybrikApplicationReconciler) validateApp(ctx context.Context, applicationContext ApplicationContext) error {
	observedStatus := applicationContext.Application.Status
	appVersion := applicationContext.Application.GetGeneration()

	// check if webhooks are enabled and application has been validated before
	// or if validated application is outdated
	if os.Getenv("ENABLE_WEBHOOKS") != "true" &&
		(string(observedStatus.ValidApplication) == "" || observedStatus.ValidatedGeneration != appVersion) {
		// do validation on applicationContext
		err := applicationContext.Application.ValidateFybrikApplication(ApplicationTaxonomy)
		applicationContext.Log.Debug().Msg("Reconciler validating Fybrik application")
		applicationContext.Application.Status.ValidatedGeneration = appVersion
		// if validation fails
		if err != nil {
			// set error message
			applicationContext.Log.Error().Err(err).Bool(logging.FORUSER, true).Bool(logging.AUDIT, true).Msg("FybrikApplication valdiation failed")
			applicationContext.Application.Status.ErrorMessage = err.Error()
			applicationContext.Application.Status.ValidApplication = v1.ConditionFalse
			return utils.UpdateStatus(ctx, r.Client, applicationContext.Application, observedStatus)
		}
		applicationContext.Application.Status.ValidApplication = v1.ConditionTrue
	}
	return nil
}
