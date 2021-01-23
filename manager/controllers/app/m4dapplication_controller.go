// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/api"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	local "github.com/ibm/the-mesh-for-data/pkg/multicluster/local"

	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// OwnerLabelKey is a key to Labels map.
	// All owned resources should be labeled using this key.
	OwnerLabelKey string = "m4d.ibm.com/owner"
)

// M4DApplicationReconciler reconciles a M4DApplication object
type M4DApplicationReconciler struct {
	client.Client
	Name              string
	Log               logr.Logger
	Scheme            *runtime.Scheme
	VaultClient       *api.Client
	PolicyCompiler    pc.IPolicyCompiler
	ResourceInterface ContextInterface
	ClusterManager    multicluster.ClusterLister
}

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=blueprints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=plotters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dmodules,verbs=get;list;watch

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

// Reconcile reconciles M4DApplication CRD
// It receives M4DApplication CRD and selects the appropriate modules that will run
// The outcome is either a single Blueprint running on the same cluster or a Plotter containing multiple Blueprints that may run on different clusters
func (r *M4DApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("m4dapplication", req.NamespacedName)
	// obtain M4DApplication resource
	applicationContext := &app.M4DApplication{}
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

	// check if reconcile is required
	// reconcile is required if the spec has been changed, or the previous reconcile has failed to allocate a Blueprint or a Plotter resource
	generationComplete := r.ResourceInterface.ResourceExists(applicationContext.Status.Generated)
	if !generationComplete || observedStatus.ObservedGeneration != applicationContext.GetGeneration() {
		if result, err := r.reconcile(applicationContext); err != nil {
			// another attempt will be done
			// users should be informed in case of errors
			if !equality.Semantic.DeepEqual(&applicationContext.Status, observedStatus) {
				// ignore an update error, a new reconcile will be made in any case
				_ = r.Client.Status().Update(ctx, applicationContext)
			}
			return result, err
		}
		applicationContext.Status.ObservedGeneration = applicationContext.GetGeneration()
	} else {
		resourceStatus, err := r.ResourceInterface.GetResourceStatus(applicationContext.Status.Generated)
		if err != nil {
			return ctrl.Result{}, err
		}
		checkResourceStatus(applicationContext, resourceStatus)
	}

	// Update CRD status in case of change (other than deletion, which was handled separately)
	if !equality.Semantic.DeepEqual(&applicationContext.Status, observedStatus) && applicationContext.DeletionTimestamp.IsZero() {
		log.V(0).Info("Reconcile: Updating status for desired generation " + fmt.Sprint(applicationContext.GetGeneration()))
		if err := r.Client.Status().Update(ctx, applicationContext); err != nil {
			return ctrl.Result{}, err
		}
	}
	if hasError(applicationContext) {
		log.Info("Reconciled with errors: " + getErrorMessages(applicationContext))
	}
	return ctrl.Result{}, nil
}

func checkResourceStatus(applicationContext *app.M4DApplication, status app.ObservedState) {
	applicationContext.Status.DataAccessInstructions = ""
	applicationContext.Status.Ready = false
	if hasError(applicationContext) {
		return
	}
	if status.Error != "" {
		setCondition(applicationContext, "", status.Error, true)
		return
	}
	if status.Ready {
		applicationContext.Status.Ready = true
		applicationContext.Status.DataAccessInstructions = status.DataAccessInstructions
	}
}

// reconcileFinalizers reconciles finalizers for M4DApplication
func (r *M4DApplicationReconciler) reconcileFinalizers(applicationContext *app.M4DApplication) error {
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

func (r *M4DApplicationReconciler) deleteExternalResources(applicationContext *app.M4DApplication) error {
	// clear provisioned buckets
	key, _ := client.ObjectKeyFromObject(applicationContext)
	if err := r.FreeStorageAssets(key); err != nil {
		return err
	}

	// delete the generated resource
	if applicationContext.Status.Generated == nil {
		return nil
	}

	r.Log.V(0).Info("Reconcile: M4DApplication is deleting the generated " + applicationContext.Status.Generated.Kind)
	if err := r.ResourceInterface.DeleteResource(applicationContext.Status.Generated); err != nil {
		return err
	}
	applicationContext.Status.Generated = nil
	return nil
}

// reconcile receives either M4DApplication CRD
// or a status update from the generated resource
func (r *M4DApplicationReconciler) reconcile(applicationContext *app.M4DApplication) (ctrl.Result, error) {
	utils.PrintStructure(applicationContext.Spec, r.Log, "M4DApplication")

	// Data User created or updated the M4DApplication

	// clear status
	resetConditions(applicationContext)
	applicationContext.Status.DataAccessInstructions = ""
	applicationContext.Status.Ready = false

	key, _ := client.ObjectKeyFromObject(applicationContext)

	// clear storage assets
	// TODO: if implicit copy is still required for the same dataset, do not free the bucket
	if err := r.FreeStorageAssets(key); err != nil {
		return ctrl.Result{}, err
	}

	clusters, err := r.ClusterManager.GetClusters()
	if err != nil {
		return ctrl.Result{}, err
	}
	// create a list of requirements for creating a data flow (actions, interface to app, data format) per a single data set
	var requirements []modules.DataInfo
	for _, dataset := range applicationContext.Spec.Data {
		req := modules.DataInfo{
			DataDetails: nil,
			Credentials: nil,
			Actions:     make(map[pb.AccessOperation_AccessType]modules.Operations),
			Context:     &dataset,
		}
		if err := r.constructDataInfo(&req, applicationContext, clusters); err != nil {
			return ctrl.Result{}, err
		}
		requirements = append(requirements, req)
	}
	// check for errors
	if hasError(applicationContext) {
		return ctrl.Result{}, nil
	}

	// create a module manager that will select modules to be orchestrated based on user requirements and module capabilities
	moduleMap, err := r.GetAllModules()
	if err != nil {
		return ctrl.Result{}, err
	}
	objectKey, _ := client.ObjectKeyFromObject(applicationContext)
	moduleManager := &ModuleManager{Client: r.Client, Log: r.Log, Modules: moduleMap, Clusters: clusters, Owner: objectKey}
	instances := make([]modules.ModuleInstanceSpec, 0)
	for _, item := range requirements {
		instancesPerDataset, err := moduleManager.SelectModuleInstances(item, applicationContext)
		if err != nil {
			setCondition(applicationContext, item.Context.DataSetID, err.Error(), true)
		}
		instances = append(instances, instancesPerDataset...)
	}
	// check for errors
	if hasError(applicationContext) {
		return ctrl.Result{}, nil
	}
	// generate blueprint specifications (per cluster)
	blueprintPerClusterMap := r.GenerateBlueprints(instances, applicationContext)
	resourceRef, err := r.ResourceInterface.CreateResourceReference(applicationContext.Name, applicationContext.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	ownerRef := &app.ResourceReference{Name: applicationContext.Name, Namespace: applicationContext.Namespace}
	if err := r.ResourceInterface.CreateOrUpdateResource(ownerRef, resourceRef, blueprintPerClusterMap); err != nil {
		r.Log.V(0).Info("Error creating " + resourceRef.Kind + " : " + err.Error())
		if err.Error() == app.InvalidClusterConfiguration {
			setCondition(applicationContext, "", app.InvalidClusterConfiguration, true)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	applicationContext.Status.Generated = resourceRef
	r.Log.V(0).Info("Created " + resourceRef.Kind + " successfully!")
	return ctrl.Result{}, nil
}

func (r *M4DApplicationReconciler) constructDataInfo(req *modules.DataInfo, input *app.M4DApplication, clusters []multicluster.Cluster) error {
	datasetID := req.Context.DataSetID
	var err error
	// Call the DataCatalog service to get info about the dataset
	if err = GetConnectionDetails(req, input); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err)
	}
	// Call the CredentialsManager service to get info about the dataset
	if err = GetCredentials(req, input); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err)
	}
	// The received credentials are stored in vault
	if err = r.RegisterCredentials(req); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err)
	}

	// policies for READ and WRITE operations based on the selected workload and data requirements
	var workloadGeography string
	if workloadGeography, err = r.GetProcessingGeography(input); err != nil {
		return err
	}

	if input.Spec.Selector.WorkloadSelector.Size() > 0 {
		// workload exists
		// read policies for data that is processed in the workload geography
		if req.Actions[pb.AccessOperation_READ], err = LookupPolicyDecisions(datasetID, r.PolicyCompiler, input,
			pb.AccessOperation{Type: pb.AccessOperation_READ, Destination: workloadGeography}); err != nil {
			return AnalyzeError(input, r.Log, datasetID, err)
		}

		// write policies in case copy will be applied
		if req.Actions[pb.AccessOperation_WRITE], err = LookupPolicyDecisions(datasetID, r.PolicyCompiler, input,
			pb.AccessOperation{Type: pb.AccessOperation_WRITE, Destination: workloadGeography}); err != nil {
			return AnalyzeError(input, r.Log, datasetID, err)
		}

		if !req.Actions[pb.AccessOperation_READ].Allowed {
			setCondition(input, datasetID, req.Actions[pb.AccessOperation_READ].Message, true)
		}
	} else {
		// workload is not selected
		// if the cluster selector is non-empty, the write will be done to the specified geography
		// Otherwise, select any of the available geographies
		if input.Spec.Selector.ClusterName != "" {
			if req.Actions[pb.AccessOperation_WRITE], err = LookupPolicyDecisions(datasetID, r.PolicyCompiler, input,
				pb.AccessOperation{Type: pb.AccessOperation_WRITE, Destination: workloadGeography}); err != nil {
				return AnalyzeError(input, r.Log, datasetID, err)
			}
			if !req.Actions[pb.AccessOperation_WRITE].Allowed {
				setCondition(input, datasetID, app.WriteNotAllowed, true)
			}
		} else {
			excludedGeos := ""
			for _, cluster := range clusters {
				operation := pb.AccessOperation{Type: pb.AccessOperation_WRITE, Destination: cluster.Metadata.Region}
				if req.Actions[pb.AccessOperation_WRITE], err = LookupPolicyDecisions(datasetID, r.PolicyCompiler, input, operation); err != nil {
					return AnalyzeError(input, r.Log, datasetID, err)
				}
				if req.Actions[pb.AccessOperation_WRITE].Allowed {
					return nil // We found a geo to which we can write
				}
				if excludedGeos != "" {
					excludedGeos += ", "
				}
				excludedGeos += cluster.Metadata.Region
			}
			// We haven't found any geographies to which we are allowed to write
			setCondition(input, datasetID, "Writing to all geographies denied: "+excludedGeos, true)
		}
	}
	return nil
}

// NewM4DApplicationReconciler creates a new reconciler for M4DApplications
func NewM4DApplicationReconciler(mgr ctrl.Manager, name string, vaultClient *api.Client,
	policyCompiler pc.IPolicyCompiler, cm multicluster.ClusterLister) *M4DApplicationReconciler {
	return &M4DApplicationReconciler{
		Client:            mgr.GetClient(),
		Name:              name,
		Log:               ctrl.Log.WithName("controllers").WithName(name),
		Scheme:            mgr.GetScheme(),
		VaultClient:       vaultClient,
		PolicyCompiler:    policyCompiler,
		ResourceInterface: NewPlotterInterface(mgr.GetClient()),
		ClusterManager:    cm,
	}
}

// SetupWithManager registers M4DApplication controller
func (r *M4DApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mapFn := handler.ToRequestsFunc(
		func(a handler.MapObject) []reconcile.Request {
			labels := a.Meta.GetLabels()
			if labels == nil {
				return []reconcile.Request{}
			}
			label, ok := labels[OwnerLabelKey]
			namespaced := strings.Split(label, ".")
			if !ok || len(namespaced) != 2 {
				return []reconcile.Request{}
			}
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Name:      namespaced[1],
					Namespace: namespaced[0],
				}},
			}
		})
	return ctrl.NewControllerManagedBy(mgr).
		For(&app.M4DApplication{}).
		Watches(&source.Kind{Type: r.ResourceInterface.GetManagedObject()},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: mapFn,
			}).Complete(r)
}

// AnalyzeError analyzes whether the given error is fatal, or a retrial attempt can be made.
// Reasons for retrial can be either communication problems with external services, or kubernetes problems to perform some action on a resource.
// A retrial is achieved by returning an error to the reconcile method
func AnalyzeError(app *app.M4DApplication, log logr.Logger, assetID string, err error) error {
	errStatus, _ := status.FromError(err)
	log.V(0).Info(errStatus.Message())
	if errStatus.Code() == codes.InvalidArgument {
		setCondition(app, assetID, errStatus.Message(), true)
		return nil
	}
	setCondition(app, assetID, errStatus.Message(), false)
	return err
}

func ownerLabels(id types.NamespacedName) map[string]string {
	return map[string]string{OwnerLabelKey: id.Namespace + "." + id.Name}
}

// GetAllModules returns all CRDs of the kind M4DModule mapped by their name
func (r *M4DApplicationReconciler) GetAllModules() (map[string]*app.M4DModule, error) {
	ctx := context.Background()

	moduleMap := make(map[string]*app.M4DModule)
	var moduleList app.M4DModuleList
	if err := r.List(ctx, &moduleList); err != nil {
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

// GetProcessingGeography determines the geography of the workload cluster.
// If no workload has been specified, a local cluster is assumed.
func (r *M4DApplicationReconciler) GetProcessingGeography(applicationContext *app.M4DApplication) (string, error) {
	clusterName := applicationContext.Spec.Selector.ClusterName
	if clusterName == "" {
		localClusterManager := local.NewManager(r.Client, utils.GetSystemNamespace())
		clusters, err := localClusterManager.GetClusters()
		if err != nil || len(clusters) != 1 {
			return "", err
		}
		return clusters[0].Metadata.Region, nil
	}
	clusters, err := r.ClusterManager.GetClusters()
	if err != nil {
		return "", err
	}
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			return cluster.Metadata.Region, nil
		}
	}
	return "", errors.New("Unknown cluster: " + clusterName)
}
