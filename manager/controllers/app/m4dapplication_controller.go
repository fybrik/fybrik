// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/api"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// M4DApplicationReconciler reconciles a M4DApplication object
type M4DApplicationReconciler struct {
	client.Client
	Name           string
	Log            logr.Logger
	Scheme         *runtime.Scheme
	VaultClient    *api.Client
	PolicyCompiler pc.IPolicyCompiler
}

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=blueprints,verbs=get;list;watch;create;update;patch;delete
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
	generatedBlueprint, _ := r.getBlueprint(applicationContext)
	if generatedBlueprint == nil || observedStatus.ObservedGeneration != applicationContext.GetGeneration() {
		if result, err := r.reconcile(applicationContext); err != nil {
			return result, err
		}
		applicationContext.Status.ObservedGeneration = applicationContext.GetGeneration()
	} else {
		checkBlueprintStatus(applicationContext, generatedBlueprint)
	}

	// Update CRD status in case of change (other than deletion, which was handled separately)
	if !equality.Semantic.DeepEqual(&applicationContext.Status, observedStatus) && applicationContext.DeletionTimestamp.IsZero() {
		log.V(0).Info("Reconcile: Updating status for desired generation " + fmt.Sprint(applicationContext.GetGeneration()))
		if err := r.Client.Status().Update(ctx, applicationContext); err != nil {
			return ctrl.Result{}, err
		}
	}
	if applicationContext.Status.Error != "" {
		log.Info("Reconciled with errors: " + applicationContext.Status.Error)
	}
	// polling for blueprint status
	if !applicationContext.Status.Ready && applicationContext.Status.Error == "" {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	return ctrl.Result{}, nil
}

func checkBlueprintStatus(applicationContext *app.M4DApplication, blueprint *app.Blueprint) {
	applicationContext.Status.DataAccessInstructions = ""
	applicationContext.Status.Ready = false
	if applicationContext.Status.Error != "" {
		return
	}
	if blueprint.Status.Error != "" {
		applicationContext.Status.Error = blueprint.Status.Error
		return
	}
	if blueprint.Status.Ready {
		applicationContext.Status.Ready = true
		applicationContext.Status.DataAccessInstructions = blueprint.Status.DataAccessInstructions
	}
}

// reconcileFinalizers reconciles finalizers for M4DApplication
func (r *M4DApplicationReconciler) reconcileFinalizers(applicationContext *app.M4DApplication) error {
	// finalizer (Blueprint)
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

	// delete the blueprint
	namespace := applicationContext.Status.BlueprintNamespace
	if namespace == "" {
		return nil
	}

	r.Log.V(0).Info("Reconcile: M4DApplication is deleting the blueprint")
	if err := r.DeleteOwnedBlueprint(applicationContext); err != nil {
		return err
	}
	// delete the allocated namespace
	if err := r.DeleteNamespace(namespace); err != nil {
		return err
	}
	return nil
}

// reconcile receives either M4DApplication CRD and generates the Blueprint CRD
// or a status update from a previously created Blueprint
func (r *M4DApplicationReconciler) reconcile(applicationContext *app.M4DApplication) (ctrl.Result, error) {
	utils.PrintStructure(applicationContext.Spec, r.Log, "M4DApplication")

	// Data User created or updated the M4DApplication

	// clear status
	applicationContext.Status.Error = ""
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
		res := r.ConstructDataInfo(dataset.DataSetID, &req, applicationContext)
		if res.Requeue {
			return res, nil
		}
		requirements = append(requirements, req)
	}
	instances := r.SelectModuleInstances(requirements, applicationContext)
	// check for errors
	if applicationContext.Status.Error != "" {
		return ctrl.Result{}, nil
	}
	r.Log.V(0).Info("Creating Blueprint")
	// first, a namespace should be created
	if applicationContext.Status.BlueprintNamespace == "" {
		if err := r.CreateNamespace(applicationContext); err != nil {
			return ctrl.Result{}, err
		}
	}
	blueprintSpec := r.GenerateBlueprint(instances, applicationContext)
	r.Log.V(0).Info("Blueprint entrypoint: " + blueprintSpec.Entrypoint)

	blueprint := r.GetBlueprintSignature(applicationContext)
	if _, err := ctrl.CreateOrUpdate(context.Background(), r, blueprint, func() error {
		blueprint.Spec = *blueprintSpec
		return nil
	}); err != nil {
		r.Log.V(0).Info("Error creating blueprint: " + err.Error())
		return ctrl.Result{}, err
	}

	/* This won't work because the m4dapplication is in a different namespace
	if err := ctrl.SetControllerReference(applicationContext, blueprint, r.Scheme); err != nil {
		log.V(0).Info("Error setting M4DApplication as owner of Blueprint: " + err.Error())
		return ctrl.Result{}, err
	}
	*/
	r.Log.V(0).Info("Succeeded in creating blueprint CRD")
	return ctrl.Result{}, nil
}

// ConstructDataInfo gets a list of governance actions and data location details for the given dataset and fills the received DataInfo structure
func (r *M4DApplicationReconciler) ConstructDataInfo(datasetID string, req *modules.DataInfo, input *app.M4DApplication) ctrl.Result {
	// policies for READ operation
	if err := LookupPolicyDecisions(datasetID, r.PolicyCompiler, req, input, pb.AccessOperation_READ); err != nil {
		return SetErrorOrRetry(input, r.Log, datasetID, err, "Policy Compiler")
	}
	if !req.Actions[app.Read].Allowed {
		SetError(input, datasetID, req.Actions[app.Read].Message, "Policy Compiler")
	}
	// Call the DataCatalog service to get info about the dataset
	if err := GetConnectionDetails(datasetID, req, input); err != nil {
		return SetErrorOrRetry(input, r.Log, datasetID, err, "Catalog Connector")
	}
	// Call the CredentialsManager service to get info about the dataset
	if err := GetCredentials(datasetID, req, input); err != nil {
		return SetErrorOrRetry(input, r.Log, datasetID, err, "Credentials Manager")
	}
	// The received credentials are stored in vault
	if err := r.RegisterCredentials(req); err != nil {
		return SetErrorOrRetry(input, r.Log, datasetID, err, "Vault")
	}

	// policies for COPY operation in case copy is required
	if err := LookupPolicyDecisions(datasetID, r.PolicyCompiler, req, input, pb.AccessOperation_COPY); err != nil {
		return SetErrorOrRetry(input, r.Log, datasetID, err, "Policy Compiler")
	}
	return ctrl.Result{}
}

// NewM4DApplicationReconciler creates a new reconciler for Blueprint resources
func NewM4DApplicationReconciler(mgr ctrl.Manager, name string, vaultClient *api.Client, policyCompiler pc.IPolicyCompiler) *M4DApplicationReconciler {
	return &M4DApplicationReconciler{
		Client:         mgr.GetClient(),
		Name:           name,
		Log:            ctrl.Log.WithName("controllers").WithName(name),
		Scheme:         mgr.GetScheme(),
		VaultClient:    vaultClient,
		PolicyCompiler: policyCompiler,
	}
}

// SetupWithManager registers M4DApplication controller
func (r *M4DApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&app.M4DApplication{}).Owns(&app.Blueprint{}).Complete(r)
}

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dmodules,verbs=get;list;watch

// GetAllModules returns all CRDs of the kind M4DModule mapped by their name
func (r *M4DApplicationReconciler) GetAllModules() map[string]*app.M4DModule {
	ctx := context.Background()

	moduleMap := make(map[string]*app.M4DModule)
	var moduleList app.M4DModuleList
	if err := r.List(ctx, &moduleList); err != nil {
		r.Log.V(0).Info("Error while listing modules: " + err.Error())
		return moduleMap
	}
	r.Log.Info("Listing all modules")
	for _, module := range moduleList.Items {
		r.Log.Info(module.GetName())
		moduleMap[module.Name] = module.DeepCopy()
	}
	return moduleMap
}

// SetErrorOrRetry analyzes whether the given error is fatal, or a retrial attempt can be made.
// Reasons for retrial can be either communication problems with external services, or kubernetes problems to perform some action on a resource.
// If an error is fatal, or a number of retrials reached the specified limit, the error is reported to the data user.
// Otherwise the error is printed to the log, and the number of retrials is increased.
func SetErrorOrRetry(app *app.M4DApplication, log logr.Logger, assetID string, err error, receivedFrom string) ctrl.Result {
	errStatus, _ := status.FromError(err)
	log.V(0).Info(errStatus.Message())
	if errStatus.Code() != codes.InvalidArgument {
		//attempt a retrial
		app.Status.NumRetries++
		if app.Status.NumRetries <= app.Spec.MaxRetries {
			return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}
		}
	}
	return SetError(app, assetID, errStatus.Message(), receivedFrom)
}

// SetError sets an error string in m4dapplication status
func SetError(app *app.M4DApplication, assetID string, msg string, receivedFrom string) ctrl.Result {
	var errorMsg string
	if assetID != "" {
		errorMsg += "Asset is " + assetID + " ; "
	}
	if receivedFrom != "" {
		errorMsg += "From " + receivedFrom + " ; "
	}
	errorMsg += msg
	if !strings.Contains(app.Status.Error, errorMsg) {
		app.Status.Error += errorMsg + "\n"
	}
	return ctrl.Result{}
}

// CreateNamespace creates a namespace in which the blueprint and the relevant resources will be running
// It stores the generated namespace name inside app status
func (r *M4DApplicationReconciler) CreateNamespace(app *app.M4DApplication) error {
	genNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{GenerateName: "m4d-"}}
	genNamespace.Labels = map[string]string{
		"m4d.ibm.com.owner": app.Namespace + "." + app.Name,
	}
	if err := r.Create(context.Background(), genNamespace); err != nil {
		return err
	}
	r.Log.V(0).Info("Created namespace " + genNamespace.Name + " for " + app.Namespace + "/" + app.Name)
	app.Status.BlueprintNamespace = genNamespace.Name
	return nil
}

// DeleteNamespace deletes the blueprint namespace upon blueprint deletion
func (r *M4DApplicationReconciler) DeleteNamespace(name string) error {
	return r.Delete(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		}})
}
