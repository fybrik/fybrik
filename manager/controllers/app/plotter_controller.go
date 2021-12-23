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

	"fybrik.io/fybrik/manager/controllers"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"github.com/rs/zerolog"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
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
	Log            zerolog.Logger
	Scheme         *runtime.Scheme
	ClusterManager multicluster.ClusterManager
}

// Reconcile receives a Plotter CRD
//nolint:dupl
func (r *PlotterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	sublog := r.Log.With().Str(logging.CONTROLLER, "Plotter").Str("plotter", req.NamespacedName.String()).Logger()

	plotter := app.Plotter{}
	if err := r.Get(ctx, req.NamespacedName, &plotter); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.reconcileFinalizers(&plotter); err != nil {
		sublog.Error().Err(err).Msg("Could not reconcile finalizers ")
		return ctrl.Result{}, err
	}

	uuid := utils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	log := sublog.With().Str(utils.FybrikAppUUID, uuid).Logger()

	// If the object has a scheduled deletion time, update status and return
	if !plotter.DeletionTimestamp.IsZero() {
		// The object is being deleted
		log.Trace().Str(logging.ACTION, logging.DELETE).Msg("Reconcile: Deleting Plotter " + plotter.GetName())
		return ctrl.Result{}, nil
	}

	observedStatus := plotter.Status.DeepCopy()
	log.Trace().Str(logging.ACTION, logging.CREATE).Msg("Reconcile: Installing/Updating Plotter " + plotter.GetName())

	result, reconcileErrors := r.reconcile(&plotter)

	if !equality.Semantic.DeepEqual(&plotter.Status, observedStatus) {
		if err := r.Status().Update(ctx, &plotter); err != nil {
			return ctrl.Result{}, errors.WrapWithDetails(err, "failed to update plotter status", "status", plotter.Status)
		}
	}

	if reconcileErrors != nil {
		log.Error().Msg("Returning with errors") // TODO - return result?
		//		log.Info("returning with errors", "result", result)
		for _, s := range reconcileErrors {
			log.Error().Err(s).Msg("Error:")
		}
		return ctrl.Result{}, errors.Wrap(reconcileErrors[0], "failed to reconcile plotter")
	}

	log.Trace().Msg("Plotter reconcile completed") // TODO - Return result?
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
	VaultAuthPath   string
	AssetID         string
	ModuleName      string
	ModuleArguments *app.StepParameters
	FlowType        app.DataFlow
	Chart           app.ChartSpec
	Scope           app.CapabilityScope
}

// addCredentials updates Vault credentials field to hold only credentials related to the flow type
func addCredentials(dataStore *app.DataStore, vaultAuthPath string, flowType app.DataFlow) {
	vaultMap := make(map[string]app.Vault)

	// Update vaultAuthPath from the cluster metadata
	// Get only flowType related creds
	vaultMap[string(flowType)] = app.Vault{
		Role:       dataStore.Vault[string(flowType)].Role,
		Address:    dataStore.Vault[string(flowType)].Address,
		SecretPath: dataStore.Vault[string(flowType)].SecretPath,
		AuthPath:   vaultAuthPath,
	}

	dataStore.Vault = vaultMap
}

// convertPlotterModuleToBlueprintModule converts an object of type PlotterModulesSpec to type ModuleInstanceSpec
func (r *PlotterReconciler) convertPlotterModuleToBlueprintModule(plotter *app.Plotter, plotterModule PlotterModulesSpec) *ModuleInstanceSpec {
	assetIDs := []string{plotterModule.AssetID}
	blueprintModule := &ModuleInstanceSpec{
		Chart:    &plotterModule.Chart,
		AssetIDs: assetIDs,
		Args: &app.ModuleArguments{
			Labels:      plotter.Labels,
			AppSelector: plotter.Spec.Selector.WorkloadSelector,
			Copy:        nil,
			Read:        nil,
			Write:       nil,
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
			addCredentials(dataStore, plotterModule.VaultAuthPath, app.ReadFlow)
		} else {
			// Fill in the DataSource from the step arguments
			dataStore = &app.DataStore{
				Connection: connectionFromService(plotterModule.ModuleArguments.Source.API),
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
		// Get only the writeFlow related creds
		// Update vaultAuthPath from the cluster metadata
		destDataStore := plotter.Spec.Assets[plotterModule.ModuleArguments.Sink.AssetID].DataStore
		addCredentials(&destDataStore, plotterModule.VaultAuthPath, app.WriteFlow)

		blueprintModule.Args.Write = []app.WriteModuleArgs{
			{
				Destination:     destDataStore,
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
			addCredentials(dataStore, plotterModule.VaultAuthPath, app.ReadFlow)
		} else {
			// Fill in the DataSource from the step arguments
			dataStore = &app.DataStore{
				Connection: connectionFromService(plotterModule.ModuleArguments.Source.API),
				Format:     plotterModule.ModuleArguments.Source.API.Format,
			}
		}
		// Get only the writeFlow related creds
		// Update vaultAuthPath from the cluster metadata
		destDataStore := plotter.Spec.Assets[plotterModule.ModuleArguments.Sink.AssetID].DataStore
		addCredentials(&destDataStore, plotterModule.VaultAuthPath, app.WriteFlow)
		blueprintModule.Args.Copy =
			&app.CopyModuleArgs{
				Source:          *dataStore,
				Destination:     destDataStore,
				AssetID:         plotterModule.AssetID,
				Transformations: plotterModule.ModuleArguments.Actions,
			}
	}
	return blueprintModule
}

// getBlueprintsMap constructs a map of blueprints driven by the plotter structure.
// The key is the cluster name.
func (r *PlotterReconciler) getBlueprintsMap(plotter *app.Plotter) map[string]app.BlueprintSpec {
	uuid := utils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	log := r.Log.With().Str(logging.CONTROLLER, "Plotter").Str(utils.FybrikAppUUID, uuid).Logger()

	log.Trace().Msg("Constructing Blueprints from Plotter")
	moduleInstances := make([]ModuleInstanceSpec, 0)

	clusters, _ := r.ClusterManager.GetClusters()

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
						scope := module.Scope
						clusterName := seqStep.Cluster
						var authPath string
						for _, cluster := range clusters {
							if clusterName == cluster.Name {
								authPath = utils.GetAuthPath(cluster.Metadata.VaultAuthPath)
								break
							}
						}
						plotterModule := PlotterModulesSpec{
							ModuleArguments: moduleArgs,
							AssetID:         flow.AssetID,
							FlowType:        subFlow.FlowType,
							ClusterName:     clusterName,
							Chart:           module.Chart,
							ModuleName:      module.Name,
							Scope:           scope,
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
			assetToStatusMap[assetID] = app.ObservedState{
				Ready: false,
				Error: err,
			}
		}
	}
}

func (r *PlotterReconciler) reconcile(plotter *app.Plotter) (ctrl.Result, []error) {
	uuid := utils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	log := r.Log.With().Str(logging.CONTROLLER, "Plotter").Str(utils.FybrikAppUUID, uuid).Logger()

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
		log.Trace().Msg("Handling spec for cluster " + cluster)
		if blueprint, exists := plotter.Status.Blueprints[cluster]; exists {
			log.Trace().Msg("Found status for cluster " + cluster)

			remoteBlueprint, err := r.ClusterManager.GetBlueprint(cluster, blueprint.Namespace, blueprint.Name)
			if err != nil {
				log.Error().Err(err).Msg("Could not fetch blueprint named " + blueprint.Name)
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
				log.Warn().Msg("Could not yet find remote blueprint")
				isReady = false
				r.setPlotterAssetsReadyStateToFalse(assetToStatusMap, &blueprintSpec, "Could not yet find remote blueprint")
				continue // Continue with next blueprint
			}

			logging.LogStructure("Remote blueprint", remoteBlueprint, log, false, true)

			if !reflect.DeepEqual(blueprintSpec, remoteBlueprint.Spec) {
				r.Log.Warn().Msg("Blueprint specs differ.  plotter.generation " + fmt.Sprint(plotter.Generation) + " plotter.observedGeneration " + fmt.Sprint(plotter.Status.ObservedGeneration))
				if plotter.Generation != plotter.Status.ObservedGeneration {
					log.Trace().Str(logging.ACTION, logging.UPDATE).Msg("Updating blueprint...")
					remoteBlueprint.Spec = blueprintSpec
					remoteBlueprint.ObjectMeta.Annotations = map[string]string(nil) // reset annotations
					err := r.ClusterManager.UpdateBlueprint(cluster, remoteBlueprint)
					if err != nil {
						log.Error().Err(err).Msg("Could not update blueprint")
						logging.LogStructure("blueprint spec", blueprintSpec, log, false, false)
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
				log.Trace().Msg("Not updating blueprint as generation did not change")
				isReady = false
				r.updatePlotterAssetsState(assetToStatusMap, remoteBlueprint)
				continue
			}

			logging.LogStructure("Remote blueprint status", remoteBlueprint.Status, log, false, false)

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
			log.Warn().Msg("Found no status for cluster " + cluster)
			blueprint := &app.Blueprint{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Blueprint",
					APIVersion: "app.fybrik.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        plotter.Name,
					Namespace:   plotter.Namespace,
					ClusterName: cluster,
					Labels: map[string]string{
						"razee/watch-resource":        "debug",
						app.ApplicationNameLabel:      plotter.Labels[app.ApplicationNameLabel],
						app.ApplicationNamespaceLabel: plotter.Labels[app.ApplicationNamespaceLabel],
					},
					Annotations: map[string]string{
						utils.FybrikAppUUID: uuid, // Pass on the globally unique id of the fybrikapplication instance for logging purposes
					},
				},
				Spec: blueprintSpec,
			}

			err := r.ClusterManager.CreateBlueprint(cluster, blueprint)
			if err != nil {
				errorCollection = append(errorCollection, err)
				log.Error().Err(err).Str(logging.CLUSTER, cluster).Str(logging.ACTION, logging.CREATE).Msg("Could not create blueprint for cluster")
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
					log.Error().Err(err).Str(logging.CLUSTER, cluster).Str(logging.BLUEPRINT, remoteBlueprint.Name).Str(logging.ACTION, logging.DELETE).Msg("Could not delete remote blueprint after spec changed!")
					continue
				}
			}
			delete(plotter.Status.Blueprints, cluster)
			log.Trace().Str(logging.PLOTTER, plotter.Name).Str(logging.CLUSTER, cluster).Str(logging.NAMESPACE, remoteBlueprint.Namespace).Str(logging.BLUEPRINT, remoteBlueprint.Name).Msg("Successfully removed blueprint from plotter")
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
			log.Trace().Str(logging.PLOTTER, plotter.Name).Msg("Plotter is ready!")
			return ctrl.Result{}, nil
		}

		// The following does a simple exponential backoff with a minimum of 5 seconds
		// and a maximum of 60 seconds until the next reconcile
		ready := *plotter.Status.ReadyTimestamp
		elapsedTime := time.Since(ready.Time)
		backoffFactor := int(math.Min(math.Exp2(elapsedTime.Minutes()), 60.0))
		requeueAfter := time.Duration(4+backoffFactor) * time.Second

		log.Trace().Str(logging.PLOTTER, plotter.Name).Str("BackoffFactor", fmt.Sprint(backoffFactor)).Str(logging.RESPONSETIME, elapsedTime.String()).Msg("Plotter is ready!")

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
		Log:            logging.LogInit(logging.CONTROLLER, name),
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

// TODO(roee88): this is temporary until we fix the API structures to use Connection properly
func connectionFromService(service *app.Service) taxonomy.Connection {
	properties := serde.Properties{}
	properties.Items["endpoint"] = service
	return taxonomy.Connection{
		Name:                 "endpoint",
		AdditionalProperties: properties,
	}
}
