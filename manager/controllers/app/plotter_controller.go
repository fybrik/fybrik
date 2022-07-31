// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
)

const (
	PlotterKind          string = "Plotter"
	PlotterFinalizerName string = "Plotter.finalizer"
)

// PlotterReconciler reconciles a Plotter object
type PlotterReconciler struct {
	client.Client
	Name           string
	Log            zerolog.Logger
	Scheme         *runtime.Scheme
	ClusterManager multicluster.ClusterManager
}

// Reconcile receives a Plotter CRD
func (r *PlotterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	sublog := r.Log.With().Str(logging.CONTROLLER, PlotterKind).Str("plotter", req.NamespacedName.String()).Logger()

	plotter := api.Plotter{}
	if err := r.Get(ctx, req.NamespacedName, &plotter); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	uuid := utils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	log := sublog.With().Str(utils.FybrikAppUUID, uuid).Logger()

	// If the object has a scheduled deletion time, update status and return
	if !plotter.DeletionTimestamp.IsZero() {
		// The object is being deleted
		log.Trace().Str(logging.ACTION, logging.DELETE).Msg("Reconcile: Deleting Plotter " + plotter.GetName())
		return ctrl.Result{}, r.removeFinalizers(ctx, &plotter)
	}

	observedStatus := plotter.Status.DeepCopy()
	log.Trace().Str(logging.ACTION, logging.CREATE).Msg("Reconcile: Installing/Updating Plotter " + plotter.GetName())

	result, reconcileErrors := r.reconcile(&plotter)
	if err := utils.UpdateStatus(ctx, r.Client, &plotter, observedStatus); err != nil {
		return ctrl.Result{}, errors.WrapWithDetails(err, "failed to update plotter status", "plotterStatus", plotter.Status)
	}

	if reconcileErrors != nil {
		log.Error().Msg("Returning with errors") // TODO - return result?	log.Info("returning with errors", "result", result)
		for _, s := range reconcileErrors {
			log.Error().Err(s).Msg("Error:")
		}
		return ctrl.Result{}, errors.Wrap(reconcileErrors[0], "failed to reconcile plotter")
	}

	log.Trace().Msg("Plotter reconcile completed") // TODO - Return result?
	return result, nil
}

// removeFinalizers removes finalizers for Plotter and deletes allocated resources
func (r *PlotterReconciler) removeFinalizers(ctx context.Context, plotter *api.Plotter) error {
	if ctrlutil.ContainsFinalizer(plotter, PlotterFinalizerName) {
		original := plotter.DeepCopy()
		// the finalizer is present - delete the allocated resources
		for cluster, blueprint := range plotter.Status.Blueprints {
			// TODO Check namespace deletion. Some finalizers leave namespaces in terminating state
			err := r.ClusterManager.DeleteBlueprint(cluster, blueprint.Namespace, blueprint.Name)
			if err != nil && client.IgnoreNotFound(err) != nil {
				return err
			}
			delete(plotter.Status.Blueprints, cluster)
		}
		// remove the finalizer from the list and update it, because it needs to be deleted together with the object
		ctrlutil.RemoveFinalizer(plotter, PlotterFinalizerName)
		if err := r.Patch(ctx, plotter, client.MergeFrom(original)); err != nil {
			return client.IgnoreNotFound(err)
		}
	}
	return nil
}

// PlotterModulesSpec consists of module details extracted from the Plotter structure
type PlotterModulesSpec struct {
	ClusterName     string
	VaultAuthPath   string
	AssetID         string
	ModuleName      string
	ModuleArguments *api.StepParameters
	FlowType        taxonomy.DataFlow
	Chart           api.ChartSpec
	Scope           api.CapabilityScope
	Capability      taxonomy.Capability
}

// addCredentials updates Vault credentials field to hold only credentials related to the flow type
func addCredentials(dataStore *api.DataStore, vaultAuthPath string, flowType taxonomy.DataFlow) {
	vaultMap := make(map[string]api.Vault)

	// Update vaultAuthPath from the cluster metadata
	// Get only flowType related creds
	vaultMap[string(flowType)] = api.Vault{
		Role:       dataStore.Vault[string(flowType)].Role,
		Address:    dataStore.Vault[string(flowType)].Address,
		SecretPath: dataStore.Vault[string(flowType)].SecretPath,
		AuthPath:   vaultAuthPath,
	}

	dataStore.Vault = vaultMap
}

// convertPlotterModuleToBlueprintModule converts an object of type PlotterModulesSpec to type ModuleInstanceSpec
func (r *PlotterReconciler) convertPlotterModuleToBlueprintModule(plotter *api.Plotter,
	plotterModule *PlotterModulesSpec) *ModuleInstanceSpec {
	blueprintModule := &ModuleInstanceSpec{
		Module: api.BlueprintModule{
			Name:  plotterModule.ModuleName,
			Chart: plotterModule.Chart,
			Arguments: api.ModuleArguments{
				Assets: []api.AssetContext{},
			},
			AssetIDs: []string{plotterModule.AssetID},
		},
		ClusterName: plotterModule.ClusterName,
		Scope:       plotterModule.Scope,
	}

	if plotterModule.ModuleArguments == nil {
		return blueprintModule
	}

	var dataStore *api.DataStore
	var destDataStore *api.DataStore
	if len(plotterModule.ModuleArguments.Arguments) > 0 && plotterModule.ModuleArguments.Arguments[0] != nil {
		if plotterModule.ModuleArguments.Arguments[0].AssetID != "" {
			assetID := plotterModule.ModuleArguments.Arguments[0].AssetID
			// Get the first argument from plotter assetID list
			assetInfo := plotter.Spec.Assets[assetID]
			dataStore = &assetInfo.DataStore
			// Get the operation of the first argument from the flow type.
			operation := plotterModule.FlowType
			if plotterModule.FlowType == taxonomy.CopyFlow {
				operation = taxonomy.ReadFlow
			}
			addCredentials(dataStore, plotterModule.VaultAuthPath, operation)
		} else {
			// Fill in the DataSource from the step arguments
			dataStore = &api.DataStore{
				Connection: plotterModule.ModuleArguments.Arguments[0].API.Connection,
				Format:     plotterModule.ModuleArguments.Arguments[0].API.DataFormat,
			}
		}
	}
	if len(plotterModule.ModuleArguments.Arguments) > 1 && plotterModule.ModuleArguments.Arguments[1] != nil {
		// Update vaultAuthPath from the cluster metadata
		assetID := plotterModule.ModuleArguments.Arguments[1].AssetID
		assetInfo := plotter.Spec.Assets[assetID]
		destDataStore = &assetInfo.DataStore
		// Get the operation of the second argument.
		// Currently it is only used in the copy flow where the second argument
		// holds information about the asset to write.
		addCredentials(destDataStore, plotterModule.VaultAuthPath, taxonomy.WriteFlow)
	}
	var args []*api.DataStore
	if dataStore != nil {
		args = append(args, dataStore)
	}
	if destDataStore != nil {
		args = append(args, destDataStore)
	}
	blueprintModule.Module.Arguments.Assets = []api.AssetContext{
		{
			Arguments:       args,
			AssetID:         plotterModule.AssetID,
			Transformations: plotterModule.ModuleArguments.Actions,
			Capability:      plotterModule.Capability,
		},
	}
	return blueprintModule
}

// getBlueprintsMap constructs a map of blueprints driven by the plotter structure.
// The key is the cluster name.
func (r *PlotterReconciler) getBlueprintsMap(plotter *api.Plotter) map[string]api.BlueprintSpec {
	uuid := utils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	log := r.Log.With().Str(logging.CONTROLLER, PlotterKind).Str(utils.FybrikAppUUID, uuid).Logger()

	log.Trace().Msg("Constructing Blueprints from Plotter")
	moduleInstances := make([]ModuleInstanceSpec, 0)

	clusters, _ := r.ClusterManager.GetClusters()

	for _, flow := range plotter.Spec.Flows {
		for _, subFlow := range flow.SubFlows {
			for _, subFlowStep := range subFlow.Steps {
				for _, seqStep := range subFlowStep {
					stepTemplate := plotter.Spec.Templates[seqStep.Template]
					for _, module := range stepTemplate.Modules {
						moduleArgs := seqStep.Parameters

						// If the module type is "plugin" then it is assumed
						// that there is a primary module of type "config" or "service"
						// in the same template and all the module arguments are used only by
						// the primary module
						if module.Type == "plugin" {
							moduleArgs = nil
						}
						scope := module.Scope
						clusterName := seqStep.Cluster
						var authPath string
						for _, cluster := range clusters {
							if clusterName == cluster.Name {
								authPath = utils.GetAuthPath(cluster.Metadata.VaultAuthPath)
								break
							}
						}
						plotterModule := &PlotterModulesSpec{
							ModuleArguments: moduleArgs,
							AssetID:         flow.AssetID,
							FlowType:        subFlow.FlowType,
							ClusterName:     clusterName,
							Chart:           module.Chart,
							ModuleName:      module.Name,
							Scope:           scope,
							Capability:      module.Capability,
							VaultAuthPath:   authPath,
						}

						blueprintModule := r.convertPlotterModuleToBlueprintModule(plotter, plotterModule)
						// append the module to the modules list
						moduleInstances = append(moduleInstances, *blueprintModule)
					}
				}
			}
		}
	}
	blueprints := r.GenerateBlueprints(moduleInstances, plotter)

	return blueprints
}

// updatePlotterAssetsState updates the status of the assets processed by the blueprint modules.
func (r *PlotterReconciler) updatePlotterAssetsState(assetToStatusMap map[string]api.ObservedState, blueprint *api.Blueprint) {
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
				assetToStatusMap[assetID] = api.ObservedState{
					Ready: false,
					Error: errMsg,
				}
			}
		}
	}
}

// setPlotterAssetsReadyStateToFalse sets to false the status of the assets processed by the blueprint modules.
func (r *PlotterReconciler) setPlotterAssetsReadyStateToFalse(assetToStatusMap map[string]api.ObservedState,
	blueprintSpec *api.BlueprintSpec, errMsg string) {
	for _, module := range blueprintSpec.Modules {
		for _, assetID := range module.AssetIDs {
			var err = errMsg
			assetToStatusMap[assetID] = api.ObservedState{
				Ready: false,
				Error: err,
			}
		}
	}
}

//nolint:funlen,gocyclo
func (r *PlotterReconciler) reconcile(plotter *api.Plotter) (ctrl.Result, []error) {
	uuid := utils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	log := r.Log.With().Str(utils.FybrikAppUUID, uuid).Logger()

	if plotter.Status.Blueprints == nil {
		plotter.Status.Blueprints = make(map[string]api.MetaBlueprint)
	}
	logging.LogStructure("reconciling Plotter...", plotter, &log, zerolog.DebugLevel, false, true)
	// Reset Assets state
	assetToStatusMap := make(map[string]api.ObservedState)
	plotter.Status.ObservedState.Error = "" // Reset error state
	// Reconciliation loop per cluster
	isReady := true

	blueprintsMap := r.getBlueprintsMap(plotter)

	var errorCollection []error
	noRemoteBlueprintWarnMsg := "Could not yet find remote blueprint"
	for cluster := range blueprintsMap {
		blueprintSpec := blueprintsMap[cluster]
		log.Trace().Msg("Handling spec for cluster " + cluster)
		if blueprint, exists := plotter.Status.Blueprints[cluster]; exists {
			log.Trace().Msg("Found status for cluster " + cluster)

			remoteBlueprint, err := r.ClusterManager.GetBlueprint(cluster, blueprint.Namespace, blueprint.Name)
			if err != nil {
				log.Error().Err(err).Msg("Could not fetch blueprint named " + blueprint.Name)
				errorCollection = append(errorCollection, err)
				// a problem to fetch a blueprint, will retry
				return ctrl.Result{}, errorCollection
			}

			if remoteBlueprint == nil {
				log.Warn().Msg(noRemoteBlueprintWarnMsg)
				isReady = false
				r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, noRemoteBlueprintWarnMsg)
				continue // Continue with next blueprint
			}

			logging.LogStructure("Remote blueprint", remoteBlueprint, &log, zerolog.DebugLevel, false, true)

			if !reflect.DeepEqual(&blueprintSpec, &remoteBlueprint.Spec) {
				r.Log.Warn().Msg("Blueprint specs differ.  plotter.generation " + fmt.Sprint(plotter.Generation) +
					" plotter.observedGeneration " + fmt.Sprint(plotter.Status.ObservedGeneration))
				if plotter.Generation != plotter.Status.ObservedGeneration {
					log.Trace().Str(logging.ACTION, logging.UPDATE).Msg("Updating blueprint...")
					remoteBlueprint.Spec = blueprintSpec
					remoteBlueprint.ObjectMeta.Annotations = map[string]string(nil) // reset annotations
					err := r.ClusterManager.UpdateBlueprint(cluster, remoteBlueprint)
					if err != nil {
						log.Error().Err(err).Msg("Could not update blueprint")
						logging.LogStructure("blueprint spec", blueprintSpec, &log, zerolog.DebugLevel, false, false)
						errorCollection = append(errorCollection, err)
						isReady = false
						r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, err.Error())
						continue
					}
					// Update meta blueprint without state as changes occur
					plotter.Status.Blueprints[cluster] = api.CreateMetaBlueprintWithoutState(remoteBlueprint)
					// Plotter cannot be ready if changes were just applied
					isReady = false
					r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, "Blueprint changes just applied")
					continue // Continue with next blueprint
				}
				log.Trace().Msg("Not updating blueprint as generation did not change")
				isReady = false
				r.updatePlotterAssetsState(assetToStatusMap, remoteBlueprint)
				continue
			}

			logging.LogStructure("Remote blueprint status", remoteBlueprint.Status, &log, zerolog.DebugLevel, false, false)

			plotter.Status.Blueprints[cluster] = api.CreateMetaBlueprint(remoteBlueprint)

			if !remoteBlueprint.Status.ObservedState.Ready {
				isReady = false
			}

			// If Blueprint has an error set it as status of plotter
			if remoteBlueprint.Status.ObservedState.Error != "" {
				plotter.Status.ObservedState.Error = remoteBlueprint.Status.ObservedState.Error
			}
			r.updatePlotterAssetsState(assetToStatusMap, remoteBlueprint)
		} else {
			log.Warn().Msg("Found no status for cluster " + cluster)
			blueprint := &api.Blueprint{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Blueprint",
					APIVersion: "app.fybrik.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        plotter.Name,
					Namespace:   plotter.Namespace,
					ClusterName: cluster,
					Labels: map[string]string{
						"razee/watch-resource": "debug",
					},
					Annotations: map[string]string{
						utils.FybrikAppUUID: uuid, // Pass on the globally unique id of the fybrikapplication instance for logging purposes
					},
				},
				Spec: blueprintSpec,
			}
			for key, val := range plotter.Labels {
				blueprint.Labels[key] = val
			}
			ctrlutil.AddFinalizer(blueprint, BlueprintFinalizerName)
			err := r.ClusterManager.CreateBlueprint(cluster, blueprint)
			isReady = false
			if err != nil {
				errorCollection = append(errorCollection, err)
				log.Error().Err(err).Str(logging.CLUSTER, cluster).Str(logging.ACTION, logging.CREATE).Msg("Could not create blueprint for cluster")
				r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, err.Error())
				continue
			}

			plotter.Status.Blueprints[cluster] = api.CreateMetaBlueprintWithoutState(blueprint)
			r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, "")
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
					log.Error().Err(err).Str(logging.CLUSTER, cluster).Str(logging.BLUEPRINT, remoteBlueprint.Name).
						Str(logging.ACTION, logging.DELETE).Msg("Could not delete remote blueprint after spec changed!")
					continue
				}
			}
			delete(plotter.Status.Blueprints, cluster)
			log.Trace().Str(logging.PLOTTER, plotter.Name).Str(logging.CLUSTER, cluster).Str(logging.NAMESPACE, remoteBlueprint.Namespace).
				Str(logging.BLUEPRINT, remoteBlueprint.Name).Msg("Successfully removed blueprint from plotter")
		}
	}

	// Update observed generation
	plotter.Status.ObservedGeneration = plotter.ObjectMeta.Generation
	plotter.Status.ObservedState.Ready = isReady
	plotter.Status.Assets = assetToStatusMap
	plotterReadyMsg := "Plotter is ready!"
	if isReady {
		if plotter.Status.ReadyTimestamp == nil {
			now := metav1.NewTime(time.Now())
			plotter.Status.ReadyTimestamp = &now
		}

		if errorCollection == nil {
			log.Trace().Str(logging.PLOTTER, plotter.Name).Msg(plotterReadyMsg)
			return ctrl.Result{}, nil
		}
		// could not remove old blueprints, will retry
		// The following does a simple exponential backoff with a minimum of 5 seconds
		// and a maximum of 60 seconds until the next reconcile
		ready := *plotter.Status.ReadyTimestamp
		elapsedTime := time.Since(ready.Time)
		backoffFactor := int(math.Min(math.Exp2(elapsedTime.Minutes()), controllers.MaximumSecondsUntillReconcile))
		requeueAfter := time.Duration(4+backoffFactor) * time.Second //nolint:revive,gomnd

		log.Trace().Str(logging.PLOTTER, plotter.Name).Str("BackoffFactor", fmt.Sprint(backoffFactor)).
			Str(logging.RESPONSETIME, elapsedTime.String()).Msg(plotterReadyMsg)

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
	// plotter is not ready
	if r.ClusterManager.IsMultiClusterSetup() {
		// TODO Once a better notification mechanism exists in razee switch to that
		return ctrl.Result{RequeueAfter: 5 * time.Second}, errorCollection //nolint:revive // for magic numbers
	}
	// don't do polling for a single cluster setup, retry in case of errors
	return ctrl.Result{}, errorCollection
}

// NewPlotterReconciler creates a new reconciler for Plotter resources
func NewPlotterReconciler(mgr ctrl.Manager, name string, manager multicluster.ClusterManager) *PlotterReconciler {
	return &PlotterReconciler{
		Client:         mgr.GetClient(),
		Name:           name,
		Log:            logging.LogInit(logging.CONTROLLER, name),
		Scheme:         mgr.GetScheme(),
		ClusterManager: manager,
	}
}

// SetupWithManager registers Plotter controller
func (r *PlotterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	numReconciles := environment.GetEnvAsInt(controllers.PlotterConcurrentReconcilesConfiguration,
		controllers.DefaultPlotterConcurrentReconciles)
	if r.ClusterManager.IsMultiClusterSetup() {
		return ctrl.NewControllerManagedBy(mgr).
			WithOptions(controller.Options{MaxConcurrentReconciles: numReconciles}).
			For(&api.Plotter{}).
			Complete(r)
	}
	// for a single cluster setup there is no need in polling since
	// blueprints can be watched by the plotter controller
	mapFn := func(obj client.Object) []reconcile.Request {
		if !obj.GetDeletionTimestamp().IsZero() {
			// the owned resource is deleted - no updates should be sent
			return []reconcile.Request{}
		}
		values, err := utils.StructToMap(obj)
		if err != nil {
			return []reconcile.Request{}
		}
		// don't send updates if the object was not reconciled
		if _, exists, err := unstructured.NestedFieldNoCopy(values, "status", "observedGeneration"); err != nil || !exists {
			return []reconcile.Request{}
		}
		return []reconcile.Request{
			{NamespacedName: client.ObjectKeyFromObject(obj)},
		}
	}
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: numReconciles}).
		For(&api.Plotter{}).
		Watches(&source.Kind{Type: &api.Blueprint{}},
			handler.EnqueueRequestsFromMapFunc(mapFn)).
		Complete(r)
}
