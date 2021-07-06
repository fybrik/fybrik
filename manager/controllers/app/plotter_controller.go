// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"math"
	"reflect"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
const BlueprintNamespace = "m4d-blueprints"

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

func (r *PlotterReconciler) reconcile(plotter *app.Plotter) (ctrl.Result, []error) {
	if plotter.Status.Blueprints == nil {
		plotter.Status.Blueprints = make(map[string]app.MetaBlueprint)
	}

	plotter.Status.ObservedState.Error = "" // Reset error state
	// Reconciliation loop per cluster
	isReady := true

	var errorCollection []error
	for cluster, blueprintSpec := range plotter.Spec.Blueprints {
		r.Log.V(1).Info("Handling spec for cluster " + cluster)
		if blueprint, exists := plotter.Status.Blueprints[cluster]; exists {
			r.Log.V(2).Info("Found status for cluster " + cluster)

			remoteBlueprint, err := r.ClusterManager.GetBlueprint(cluster, blueprint.Namespace, blueprint.Name)
			if err != nil {
				r.Log.Error(err, "Could not fetch blueprint", "name", blueprint.Name)
				errorCollection = append(errorCollection, err)
				isReady = false
				continue
			}

			if remoteBlueprint == nil {
				r.Log.Info("Could not yet find remote blueprint")
				isReady = false
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
						continue
					}
					// Update meta blueprint without state as changes occur
					plotter.Status.Blueprints[cluster] = app.CreateMetaBlueprintWithoutState(remoteBlueprint)
					// Plotter cannot be ready if changes were just applied
					isReady = false
					continue // Continue with next blueprint
				}
				r.Log.V(1).Info("Not updating blueprint as generation did not change")
				isReady = false
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
		} else {
			r.Log.V(2).Info("Found no status for cluster " + cluster)
			blueprint := &app.Blueprint{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Blueprint",
					APIVersion: "app.m4d.ibm.com/v1alpha1",
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
				continue
			}

			plotter.Status.Blueprints[cluster] = app.CreateMetaBlueprintWithoutState(blueprint)
			isReady = false
		}
	}

	// Tidy up blueprints that have been deployed but are not in the spec any more
	// E.g. after a plotter has been updated
	for cluster, remoteBlueprint := range plotter.Status.Blueprints {
		if _, exists := plotter.Spec.Blueprints[cluster]; !exists {
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

	if isReady {
		if plotter.Status.ReadyTimestamp == nil {
			now := metav1.NewTime(time.Now())
			plotter.Status.ReadyTimestamp = &now
		}

		aggregatedInstructions := ""
		for _, blueprint := range plotter.Status.Blueprints {
			if len(blueprint.Status.ObservedState.DataAccessInstructions) > 0 {
				aggregatedInstructions = aggregatedInstructions + blueprint.Status.ObservedState.DataAccessInstructions + "\n"
			}
		}
		plotter.Status.ObservedState.DataAccessInstructions = aggregatedInstructions

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
	// 'UpdateFunc' and 'CreateFunc' used to judge if the event came from within the system namespace.
	// If that is true, the event will be processed by the reconciler.
	// If it's not then it is a rogue event created by someone outside of the control plane.
	p := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetNamespace() == utils.GetSystemNamespace()
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectOld.GetNamespace() == utils.GetSystemNamespace()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return e.Object.GetNamespace() == utils.GetSystemNamespace()
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&app.Plotter{}).
		WithEventFilter(p).
		Complete(r)
}
