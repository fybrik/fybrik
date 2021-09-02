// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fybrik.io/fybrik/manager/controllers"
	"fybrik.io/fybrik/pkg/environment"
	"math"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"strings"
	"time"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app/modules"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PlotterReconciler reconciles a Plotter object
type PlotterReconciler struct {
	client.Client
	Name           string
	Log            logr.Logger
	Scheme         *runtime.Scheme
	ClusterManager multicluster.ClusterManager
}

// BlueprintNamespace defines a namespace where blueprints and associated resources will be allocated
const BlueprintNamespace = "fybrik-blueprints"

// Reconcile receives a Plotter CRD
//nolint:dupl
func (r *PlotterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("plotter", req.NamespacedName)

	plotter := app.Plotter{}
	if err := r.Get(ctx, req.NamespacedName, &plotter); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.reconcileFinalizers(&plotter); err != nil {
		log.V(0).Info("Could not reconcile finalizers " + err.Error())
		return ctrl.Result{}, err
	}

	// If the object has a scheduled deletion time, update status and return
	if !plotter.DeletionTimestamp.IsZero() {
		// The object is being deleted
		log.V(0).Info("Reconcile: Deleting Plotter " + plotter.GetName())
		return ctrl.Result{}, nil
	}

	observedStatus := plotter.Status.DeepCopy()
	log.V(0).Info("Reconcile: Installing/Updating Plotter " + plotter.GetName())

	result, reconcileErrors := r.reconcile(&plotter)

	if !equality.Semantic.DeepEqual(&plotter.Status, observedStatus) {
		if err := r.Status().Update(ctx, &plotter); err != nil {
			return ctrl.Result{}, errors.WrapWithDetails(err, "failed to update plotter status", "status", plotter.Status)
		}
	}

	if reconcileErrors != nil {
		log.Info("returning with errors", "result", result)
		for _, s := range reconcileErrors {
			log.Error(s, "Error:")
		}
		return ctrl.Result{}, errors.Wrap(reconcileErrors[0], "failed to reconcile plotter")
	}

	log.Info("plotter reconcile cycle completed", "result", result)
	return result, nil
}

// reconcileFinalizers reconciles finalizers for Plotter
func (r *PlotterReconciler) reconcileFinalizers(plotter *app.Plotter) error {
	// finalizer
	finalizerName := r.Name + ".finalizer"
	hasFinalizer := ctrlutil.ContainsFinalizer(plotter, finalizerName)

	// If the object has a scheduled deletion time, delete it and its associated resources
	if !plotter.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if hasFinalizer { // Finalizer was created when the object was created
			// the finalizer is present - delete the allocated resources

			for cluster, blueprint := range plotter.Status.Blueprints {
				// TODO Check namespace deletion. Some finalizers leave namespaces in terminating state
				err := r.ClusterManager.DeleteBlueprint(cluster, blueprint.Namespace, blueprint.Name)
				if err != nil {
					return err
				}
			}

			// remove the finalizer from the list and update it, because it needs to be deleted together with the object
			ctrlutil.RemoveFinalizer(plotter, finalizerName)

			if err := r.Update(context.Background(), plotter); err != nil {
				return err
			}
		}
		return nil
	}
	// Make sure this CRD instance has a finalizer
	if !hasFinalizer {
		ctrlutil.AddFinalizer(plotter, finalizerName)
		if err := r.Update(context.Background(), plotter); err != nil {
			return err
		}
	}
	return nil
}

// PlotterModulesSpec consists of module details extracted from the Plotter structure
type PlotterModulesSpec struct {
	ClusterName     string
	AssetID         string
	ModuleName      string
	ModuleArguments *app.StepParameters
	FlowType        app.DataFlow
	Chart           app.ChartSpec
	Scope           app.CapabilityScope
}

// convertPlotterModuleToBlueprintModule converts an object of type PlotterModulesSpec to type modules.ModuleInstanceSpec
func (r *PlotterReconciler) convertPlotterModuleToBlueprintModule(plotter *app.Plotter, plotterModule PlotterModulesSpec) *modules.ModuleInstanceSpec {
	assetIDs := []string{plotterModule.AssetID}
	blueprintModule := &modules.ModuleInstanceSpec{
		Chart:    &plotterModule.Chart,
		AssetIDs: assetIDs,
		Args: &app.ModuleArguments{
			Copy:  nil,
			Read:  nil,
			Write: nil,
		},
		ClusterName: plotterModule.ClusterName,
		ModuleName:  plotterModule.ModuleName,
		Scope:       plotterModule.Scope,
	}

	if plotterModule.ModuleArguments == nil {
		return blueprintModule
	}

	switch plotterModule.FlowType {
	case app.ReadFlow:
		var dataStore *app.DataStore
		if plotterModule.ModuleArguments.Source.AssetID != "" {
			assetID := plotterModule.ModuleArguments.Source.AssetID
			// Get source from plotter assetID list
			assetInfo := plotter.Spec.Assets[assetID]
			dataStore = &assetInfo.DataStore
		} else {
			// Fill in the DataSource from the step arguments
			dataStore = &app.DataStore{
				Connection: *serde.NewArbitrary(plotterModule.ModuleArguments.Source.API),
				Format:     plotterModule.ModuleArguments.Source.API.Format,
			}
		}
		blueprintModule.Args.Read = []app.ReadModuleArgs{
			{
				Source:          *dataStore,
				AssetID:         plotterModule.AssetID,
				Transformations: plotterModule.ModuleArguments.Actions,
			},
		}
	case app.WriteFlow:
		blueprintModule.Args.Write = []app.WriteModuleArgs{
			{
				Destination:     plotter.Spec.Assets[plotterModule.ModuleArguments.Sink.AssetID].DataStore,
				AssetID:         plotterModule.AssetID,
				Transformations: plotterModule.ModuleArguments.Actions,
			},
		}
	case app.CopyFlow:
		var dataStore *app.DataStore
		if plotterModule.ModuleArguments.Source.AssetID != "" {
			assetID := plotterModule.ModuleArguments.Source.AssetID
			// Get source from plotter assetID list
			assetInfo := plotter.Spec.Assets[assetID]
			dataStore = &assetInfo.DataStore
		} else {
			// Fill in the DataSource from the step arguments
			dataStore = &app.DataStore{
				Connection: *serde.NewArbitrary(plotterModule.ModuleArguments.Source.API),
				Format:     plotterModule.ModuleArguments.Source.API.Format,
			}
		}
		blueprintModule.Args.Copy =
			&app.CopyModuleArgs{
				Source:          *dataStore,
				Destination:     plotter.Spec.Assets[plotterModule.ModuleArguments.Sink.AssetID].DataStore,
				AssetID:         plotterModule.AssetID,
				Transformations: plotterModule.ModuleArguments.Actions,
			}
	}
	return blueprintModule
}

// getModuleScope returns the scope of the module given the data flow type of the module and the module
// capabilities.
func (r *PlotterReconciler) getModuleScope(capabilities []app.ModuleCapability, moduleFlow app.DataFlow) app.CapabilityScope {
	var scope app.CapabilityScope
	for _, capability := range capabilities {
		// It is assumed that all capabilities of the same type have the same scope
		if capability.Capability == app.CapabilityType(moduleFlow) {
			scope = capability.Scope
			if capability.Scope == "" {
				// If scope is not indicated it is assumed to be asset
				scope = app.Asset
			}
			return scope
		}
	}
	// ??? we should not get here
	return app.Asset
}

// getBlueprintsMap constructs a map of blueprints driven by the plotter structure.
// The key is the cluster name.
func (r *PlotterReconciler) getBlueprintsMap(plotter *app.Plotter) map[string]app.BlueprintSpec {
	r.Log.V(1).Info("Constructing Blueprints from Plotter")
	moduleInstances := make([]modules.ModuleInstanceSpec, 0)

	for _, flow := range plotter.Spec.Flows {
		for _, subFlow := range flow.SubFlows {
			for _, subFlowStep := range subFlow.Steps {
				for _, seqStep := range subFlowStep {
					stepTemplate := plotter.Spec.Templates[seqStep.Template]
					// isPrimaryModule := true
					for _, module := range stepTemplate.Modules {
						moduleArgs := seqStep.Parameters

						// If the module type is "plugin" then it is assumed
						// that there is a primary module of type "config" or "service"
						// in the same template and all the module arguments are used only by
						// the primary module
						if module.Type == "plugin" {
							moduleArgs = nil
						}
						scope := r.getModuleScope(module.Capabilities, subFlow.FlowType)
						plotterModule := PlotterModulesSpec{
							ModuleArguments: moduleArgs,
							AssetID:         flow.AssetID,
							FlowType:        subFlow.FlowType,
							ClusterName:     seqStep.Cluster,
							ModuleName:      module.Name,
							Scope:           scope,
						}
						blueprintModule := r.convertPlotterModuleToBlueprintModule(plotter, plotterModule)
						// append the module to the modules list
						moduleInstances = append(moduleInstances, *blueprintModule)
					}
				}
			}
		}
	}
	blueprints := r.GenerateBlueprints(moduleInstances)

	return blueprints
}

// updatePlotterAssetsState updates the status of the assets processed by the blueprint modules.
func (r *PlotterReconciler) updatePlotterAssetsState(assetToStatusMap map[string]app.ObservedState, blueprint *app.Blueprint) {
	for instanceName, moduleState := range blueprint.Status.ModulesState {
		for _, assetID := range blueprint.Spec.Modules[instanceName].AssetIDs {
			state, exists := assetToStatusMap[assetID]
			errMsg := moduleState.Error
			if exists {
				errMsg = state.Error + "\n" + errMsg
			}
			if !exists {
				assetToStatusMap[assetID] = moduleState
				// if the current module is not ready then update all its assets
				// to not ready state regardless of the assets state
			} else if !moduleState.Ready {
				assetToStatusMap[assetID] = app.ObservedState{
					Ready: false,
					Error: errMsg,
				}
			}
		}
	}
}

// setPlotterAssetsReadyStateToFalse sets to false the status of the assets processed by the blueprint modules.
func (r *PlotterReconciler) setPlotterAssetsReadyStateToFalse(assetToStatusMap map[string]app.ObservedState, blueprintSpec *app.BlueprintSpec, errMsg string) {
	for _, module := range blueprintSpec.Modules {
		for _, assetID := range module.AssetIDs {
			var err = errMsg
			state, exists := assetToStatusMap[assetID]
			if exists {
				err = state.Error + err
				delete(assetToStatusMap, assetID)
			}

			assetToStatusMap[assetID] = app.ObservedState{
				Ready: false,
				Error: err,
			}
		}
	}
}

func (r *PlotterReconciler) reconcile(plotter *app.Plotter) (ctrl.Result, []error) {
	if plotter.Status.Blueprints == nil {
		plotter.Status.Blueprints = make(map[string]app.MetaBlueprint)
	}

	// Reset Assets state
	assetToStatusMap := make(map[string]app.ObservedState)
	plotter.Status.ObservedState.Error = "" // Reset error state
	// Reconciliation loop per cluster
	isReady := true

	blueprintsMap := r.getBlueprintsMap(plotter)
	var errorCollection []error

	for cluster := range blueprintsMap {
		blueprintSpec := blueprintsMap[cluster]
		r.Log.V(1).Info("Handling spec for cluster " + cluster)
		if blueprint, exists := plotter.Status.Blueprints[cluster]; exists {
			r.Log.V(2).Info("Found status for cluster " + cluster)

			remoteBlueprint, err := r.ClusterManager.GetBlueprint(cluster, blueprint.Namespace, blueprint.Name)
			if err != nil {
				r.Log.Error(err, "Could not fetch blueprint", "name", blueprint.Name)
				errorCollection = append(errorCollection, err)

				// The following does a simple exponential backoff with a minimum of 5 seconds
				// and a maximum of 60 seconds until the next reconcile
				now := metav1.NewTime(time.Now())
				elapsedTime := time.Since(now.Time)
				backoffFactor := int(math.Min(math.Exp2(elapsedTime.Minutes()), 60.0))
				requeueAfter := time.Duration(4+backoffFactor) * time.Second
				return ctrl.Result{RequeueAfter: requeueAfter}, errorCollection
			}

			if remoteBlueprint == nil {
				r.Log.Info("Could not yet find remote blueprint")
				isReady = false
				r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, "Could not yet find remote blueprint")
				continue // Continue with next blueprint
			}

			r.Log.V(2).Info("Remote blueprint: ", "rbp", remoteBlueprint)

			if !reflect.DeepEqual(blueprintSpec, remoteBlueprint.Spec) {
				r.Log.V(1).Info("Blueprint specs differ",
					"plotter.generation", plotter.Generation,
					"plotter.observedGeneration", plotter.Status.ObservedGeneration)
				if plotter.Generation != plotter.Status.ObservedGeneration {
					r.Log.V(1).Info("Updating blueprint...")
					remoteBlueprint.Spec = blueprintSpec
					remoteBlueprint.ObjectMeta.Annotations = map[string]string(nil) // reset annotations
					err := r.ClusterManager.UpdateBlueprint(cluster, remoteBlueprint)
					if err != nil {
						r.Log.Error(err, "Could not update blueprint", "newSpec", blueprintSpec)
						errorCollection = append(errorCollection, err)
						isReady = false
						r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, err.Error())
						continue
					}
					// Update meta blueprint without state as changes occur
					plotter.Status.Blueprints[cluster] = app.CreateMetaBlueprintWithoutState(remoteBlueprint)
					// Plotter cannot be ready if changes were just applied
					isReady = false
					r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, "Blueprint changes just applied")
					continue // Continue with next blueprint
				}
				r.Log.V(1).Info("Not updating blueprint as generation did not change")
				isReady = false
				r.updatePlotterAssetsState(assetToStatusMap, remoteBlueprint)
				continue
			}

			r.Log.V(2).Info("Status of remote blueprint ", "status", remoteBlueprint.Status)

			plotter.Status.Blueprints[cluster] = app.CreateMetaBlueprint(remoteBlueprint)

			if !remoteBlueprint.Status.ObservedState.Ready {
				isReady = false
			}

			// If Blueprint has an error set it as status of plotter
			if remoteBlueprint.Status.ObservedState.Error != "" {
				plotter.Status.ObservedState.Error = remoteBlueprint.Status.ObservedState.Error
			}
			r.updatePlotterAssetsState(assetToStatusMap, remoteBlueprint)
		} else {
			r.Log.V(2).Info("Found no status for cluster " + cluster)
			blueprint := &app.Blueprint{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Blueprint",
					APIVersion: "app.fybrik.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        plotter.Name,
					Namespace:   BlueprintNamespace,
					ClusterName: cluster,
					Labels: map[string]string{
						"razee/watch-resource":        "debug",
						app.ApplicationNameLabel:      plotter.Labels[app.ApplicationNameLabel],
						app.ApplicationNamespaceLabel: plotter.Labels[app.ApplicationNamespaceLabel],
					},
				},
				Spec: blueprintSpec,
			}

			err := r.ClusterManager.CreateBlueprint(cluster, blueprint)
			if err != nil {
				errorCollection = append(errorCollection, err)
				r.Log.Error(err, "Could not create blueprint for cluster", "cluster", cluster)
				r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, err.Error())
				continue
			}

			plotter.Status.Blueprints[cluster] = app.CreateMetaBlueprintWithoutState(blueprint)
			isReady = false
			r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, "Blueprint just created")
		}
	}

	// Tidy up blueprints that have been deployed but are not in the spec any more
	// E.g. after a plotter has been updated
	for cluster, remoteBlueprint := range plotter.Status.Blueprints {
		if _, exists := blueprintsMap[cluster]; !exists {
			err := r.ClusterManager.DeleteBlueprint(cluster, remoteBlueprint.Namespace, remoteBlueprint.Name)
			if err != nil {
				if !strings.HasPrefix(err.Error(), "Query channelByName error. Could not find the channel with name") {
					errorCollection = append(errorCollection, err)
					r.Log.Error(err, "Could not delete remote blueprint after spec changed!", "cluster", cluster)
					continue
				}
			}
			delete(plotter.Status.Blueprints, cluster)
			r.Log.V(1).Info("Successfully removed blueprint from plotter",
				"plotter", plotter.Name,
				"cluster", cluster,
				"namespace", remoteBlueprint.Namespace,
				"name", remoteBlueprint.Name)
		}
	}

	// Update observed generation
	plotter.Status.ObservedGeneration = plotter.ObjectMeta.Generation
	plotter.Status.ObservedState.Ready = isReady
	plotter.Status.Assets = assetToStatusMap

	if isReady {
		if plotter.Status.ReadyTimestamp == nil {
			now := metav1.NewTime(time.Now())
			plotter.Status.ReadyTimestamp = &now
		}

		if errorCollection == nil {
			r.Log.V(2).Info("Plotter is ready!", "plotter", plotter.Name)
			return ctrl.Result{}, nil
		}

		// The following does a simple exponential backoff with a minimum of 5 seconds
		// and a maximum of 60 seconds until the next reconcile
		ready := *plotter.Status.ReadyTimestamp
		elapsedTime := time.Since(ready.Time)
		backoffFactor := int(math.Min(math.Exp2(elapsedTime.Minutes()), 60.0))
		requeueAfter := time.Duration(4+backoffFactor) * time.Second

		r.Log.V(2).Info("Plotter is ready!", "plotter", plotter.Name, "backoffFactor", backoffFactor, "elapsedTime", elapsedTime)

		return ctrl.Result{RequeueAfter: requeueAfter}, errorCollection
	}

	plotter.Status.ReadyTimestamp = nil

	// If no error was set from observed state set it to possible errorCollection that appeared
	if plotter.Status.ObservedState.Error == "" && errorCollection != nil {
		aggregatedError := ""
		for _, err := range errorCollection {
			aggregatedError = aggregatedError + err.Error() + "\n"
		}
		plotter.Status.ObservedState.Error = aggregatedError
	}

	// TODO Once a better notification mechanism exists in razee switch to that
	return ctrl.Result{RequeueAfter: 5 * time.Second}, errorCollection
}

// NewPlotterReconciler creates a new reconciler for Plotter resources
func NewPlotterReconciler(mgr ctrl.Manager, name string, manager multicluster.ClusterManager) *PlotterReconciler {
	return &PlotterReconciler{
		Client:         mgr.GetClient(),
		Name:           name,
		Log:            ctrl.Log.WithName("controllers").WithName(name),
		Scheme:         mgr.GetScheme(),
		ClusterManager: manager,
	}
}

// SetupWithManager registers Plotter controller
func (r *PlotterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	numReconciles := environment.GetEnvAsInt(controllers.PlotterConcurrentReconcilesConfiguration, controllers.DefaultPlotterConcurrentReconciles)

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: numReconciles}).
		For(&app.Plotter{}).
		Complete(r)
}
