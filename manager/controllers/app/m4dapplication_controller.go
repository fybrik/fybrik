// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/api"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
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
}

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=blueprints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=plotters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

// Reconcile reconciles M4DApplication CRD
// It receives M4DApplication CRD and generates the Blueprint CRD
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
	// reconcile is required if the spec has been changed, or no blueprint generation was done (reconcile due to non-fatal errors)
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
		checkBlueprintStatus(applicationContext, resourceStatus)
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
	// polling for blueprint status
	if !applicationContext.Status.Ready && !hasError(applicationContext) {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	return ctrl.Result{}, nil
}

func checkBlueprintStatus(applicationContext *app.M4DApplication, status ResourceStatus) {
	applicationContext.Status.DataAccessInstructions = ""
	applicationContext.Status.Ready = false
	if hasError(applicationContext) {
		return
	}
	if status.Error != "" {
		setCondition(applicationContext, "", status.Error, "Orchestration", true)
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

	// If the object has a scheduled deletion time, delete it and its associated blueprint
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

	r.Log.V(0).Info("Reconcile: M4DApplication is deleting the blueprint")
	if err := r.ResourceInterface.DeleteResource(applicationContext.Status.Generated); err != nil {
		return err
	}
	applicationContext.Status.Generated = nil
	return nil
}

// reconcile receives either M4DApplication CRD and generates the blueprint resource
// or a status update from a previously created blueprint resource
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

	// create a list of requirements for creating a data flow (actions, interface to app, data format) per a single data set
	// A unique identifier (AssetID) is used to represent the dataset in the internal flow (for logs, map keys, vault path creation)
	// The original dataset.DataSetID is used for communication with the connectors
	var requirements []modules.DataInfo
	for _, dataset := range applicationContext.Spec.Data {
		req := modules.DataInfo{
			AssetID:      dataset.DataSetID,
			DataDetails:  nil,
			Credentials:  nil,
			Actions:      make(map[app.ModuleFlow]modules.Transformations),
			AppInterface: &dataset.IFdetails,
		}
		// get enforcement actions and location info for a dataset
		if err := r.ConstructDataInfo(dataset.DataSetID, &req, applicationContext); err != nil {
			return ctrl.Result{}, err
		}
		requirements = append(requirements, req)
	}
	instances, err := r.SelectModuleInstances(requirements, applicationContext)
	if err != nil {
		return ctrl.Result{}, err
	}
	// check for errors
	if hasError(applicationContext) {
		return ctrl.Result{}, nil
	}
	// unite several instances of a read/write module
	newInstances := r.RefineInstances(instances)
	r.Log.V(0).Info("Creating blueprint(s)")

	blueprintSpec := r.GenerateBlueprint(newInstances, applicationContext)
	blueprintPerClusterMap := make(map[string]app.BlueprintSpec)
	blueprintPerClusterMap[applicationContext.ClusterName] = *blueprintSpec

	resourceMetadata, err := r.ResourceInterface.CreateResourceMetadata(applicationContext.Name, applicationContext.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ResourceInterface.CreateOrUpdateResource(resourceMetadata, blueprintPerClusterMap); err != nil {
		r.Log.V(0).Info("Error creating blueprint: " + err.Error())
		return ctrl.Result{}, err
	}
	applicationContext.Status.Generated = resourceMetadata
	r.Log.V(0).Info("Succeeded in blueprint creation")
	return ctrl.Result{}, nil
}

// ConstructDataInfo gets a list of governance actions and data location details for the given dataset and fills the received DataInfo structure
func (r *M4DApplicationReconciler) ConstructDataInfo(datasetID string, req *modules.DataInfo, input *app.M4DApplication) error {
	// policies for READ operation
	if err := LookupPolicyDecisions(datasetID, r.PolicyCompiler, req, input, pb.AccessOperation_READ); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err, "Policy Compiler")
	}
	if !req.Actions[app.Read].Allowed {
		setCondition(input, datasetID, req.Actions[app.Read].Message, "Policy Compiler", true)
	}
	// Call the DataCatalog service to get info about the dataset
	if err := GetConnectionDetails(datasetID, req, input); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err, "Catalog Connector")
	}
	// Call the CredentialsManager service to get info about the dataset
	if err := GetCredentials(datasetID, req, input); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err, "Credentials Manager")
	}
	// The received credentials are stored in vault
	if err := r.RegisterCredentials(req); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err, "Vault")
	}

	// policies for COPY operation in case copy is required
	if err := LookupPolicyDecisions(datasetID, r.PolicyCompiler, req, input, pb.AccessOperation_COPY); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err, "Policy Compiler")
	}
	return nil
}

// NewM4DApplicationReconciler creates a new reconciler for Blueprint resources
func NewM4DApplicationReconciler(mgr ctrl.Manager, name string, vaultClient *api.Client, policyCompiler pc.IPolicyCompiler, context ContextInterface) *M4DApplicationReconciler {
	return &M4DApplicationReconciler{
		Client:            mgr.GetClient(),
		Name:              name,
		Log:               ctrl.Log.WithName("controllers").WithName(name),
		Scheme:            mgr.GetScheme(),
		VaultClient:       vaultClient,
		PolicyCompiler:    policyCompiler,
		ResourceInterface: context,
	}
}

// SetupWithManager registers M4DApplication controller
func (r *M4DApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&app.M4DApplication{}).
		Watches(&source.Kind{Type: r.ResourceInterface.GetManagedObject()}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dmodules,verbs=get;list;watch

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

// AnalyzeError analyzes whether the given error is fatal, or a retrial attempt can be made.
// Reasons for retrial can be either communication problems with external services, or kubernetes problems to perform some action on a resource.
// A retrial is achieved by returning an error to the reconcile method
func AnalyzeError(app *app.M4DApplication, log logr.Logger, assetID string, err error, receivedFrom string) error {
	errStatus, _ := status.FromError(err)
	log.V(0).Info(errStatus.Message())
	if errStatus.Code() == codes.InvalidArgument {
		setCondition(app, assetID, errStatus.Message(), receivedFrom, true)
		return nil
	}
	setCondition(app, assetID, errStatus.Message(), receivedFrom, false)
	return err
}

// Helper functions to manage conditions
func resetConditions(application *app.M4DApplication) {
	application.Status.Conditions = make([]app.Condition, 2)
	application.Status.Conditions[app.ErrorConditionIndex] = app.Condition{Type: app.ErrorCondition, Status: corev1.ConditionFalse}
	application.Status.Conditions[app.FailureConditionIndex] = app.Condition{Type: app.FailureCondition, Status: corev1.ConditionFalse}
}

func setCondition(application *app.M4DApplication, assetID string, msg string, receivedFrom string, fatalError bool) {
	if len(application.Status.Conditions) == 0 {
		resetConditions(application)
	}
	errMsg := "An error was received"
	if receivedFrom != "" {
		errMsg += " from " + receivedFrom
	}
	if assetID != "" {
		errMsg += " for asset " + assetID + " . "
	}
	if !fatalError {
		errMsg += "If the error persists, please contact an operator.\n"
	}
	errMsg += "Error description: " + msg + "\n"
	var ind int64
	if fatalError {
		ind = app.FailureConditionIndex
	} else {
		ind = app.ErrorConditionIndex
	}
	application.Status.Conditions[ind].Status = corev1.ConditionTrue
	application.Status.Conditions[ind].Message += errMsg
}

func hasError(application *app.M4DApplication) bool {
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return false
	}
	return (application.Status.Conditions[app.ErrorConditionIndex].Status == corev1.ConditionTrue ||
		application.Status.Conditions[app.FailureConditionIndex].Status == corev1.ConditionTrue)
}

func getErrorMessages(application *app.M4DApplication) string {
	var errMsg string
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return errMsg
	}
	if application.Status.Conditions[app.ErrorConditionIndex].Status == corev1.ConditionTrue {
		errMsg += application.Status.Conditions[app.ErrorConditionIndex].Message
	}
	if application.Status.Conditions[app.FailureConditionIndex].Status == corev1.ConditionTrue {
		errMsg += application.Status.Conditions[app.FailureConditionIndex].Message
	}
	return errMsg
}
