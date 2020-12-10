// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"emperror.dev/errors"
	"github.com/go-logr/logr"
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

// PlotterReconciler reconciles a Plotter object
type PlotterReconciler struct {
	client.Client
	Name           string
	Log            logr.Logger
	Scheme         *runtime.Scheme
	ClusterManager multicluster.ClusterManager
}

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=plotters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=plotters/status,verbs=get;update;patch

// Reconcile receives a Plotter CRD
//nolint:dupl
func (r *PlotterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("plotter", req.NamespacedName)
	var err error

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

	result, err := r.reconcile(&plotter)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to reconcile plotter")
	}

	if !equality.Semantic.DeepEqual(&plotter.Status, observedStatus) {
		if err := r.Client.Status().Update(ctx, &plotter); err != nil {
			return ctrl.Result{}, errors.WrapWithDetails(err, "failed to update plotter status", "status", plotter.Status)
		}
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

func (r *PlotterReconciler) reconcile(plotter *app.Plotter) (ctrl.Result, error) {
	isInitialReconcile := false
	if plotter.Status.Blueprints == nil {
		plotter.Status.Blueprints = make(map[string]app.MetaBlueprint)
		isInitialReconcile = true
	}

	// Reconciliation loop per cluster
	isReady := true
	for cluster, blueprintSpec := range plotter.Spec.Blueprints {
		r.Log.V(1).Info("Handling cluster " + cluster)
		if blueprint, exists := plotter.Status.Blueprints[cluster]; exists {
			r.Log.V(2).Info("Found status for cluster " + cluster)

			remoteBlueprint, err := r.ClusterManager.GetBlueprint(cluster, blueprint.Namespace, blueprint.Name)
			if err != nil {
				return ctrl.Result{}, err
			}

			if remoteBlueprint == nil {
				r.Log.Info("Could not yet find remote blueprint")
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}

			if !reflect.DeepEqual(blueprintSpec, remoteBlueprint.Spec) {
				r.Log.V(1).Info("Blueprint specs differ...")
				remoteBlueprint.Spec = blueprintSpec
				remoteBlueprint.ObjectMeta.Annotations = map[string]string(nil) // reset annotations
				err := r.ClusterManager.UpdateBlueprint(cluster, remoteBlueprint)
				if err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}

			r.Log.V(2).Info("Status of remote blueprint ", "status", remoteBlueprint.Status)

			blueprint.Status = remoteBlueprint.Status

			plotter.Status.Blueprints[cluster] = blueprint

			if !remoteBlueprint.Status.ObservedState.Ready {
				isReady = false
			}

			// If Blueprint has an error set it as status of plotter
			if remoteBlueprint.Status.ObservedState.Error != "" {
				plotter.Status.ObservedState.Error = remoteBlueprint.Status.ObservedState.Error
			}
		} else {
			random := rand.String(5)
			randomNamespace := "m4d-" + random
			blueprint := &app.Blueprint{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Blueprint",
					APIVersion: "app.m4d.ibm.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "blueprint",
					Namespace:   randomNamespace,
					ClusterName: cluster,
					Labels:      map[string]string{"razee/watch-resource": "debug"},
				},
				Spec: blueprintSpec,
			}

			blueprintMini := app.MetaBlueprint{
				ObjectMeta: blueprint.ObjectMeta,
				Status:     app.BlueprintStatus{},
			}

			plotter.Status.Blueprints[cluster] = blueprintMini
			err := r.ClusterManager.CreateBlueprint(cluster, blueprint)
			if err != nil {
				return ctrl.Result{}, err
			}
			isReady = false
		}
	}

	// TODO do loop of statuses vs spec for removed specs

	if isInitialReconcile {
		// Return after initial deployment of blueprints
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if isReady {
		plotter.Status.ObservedState.Ready = true

		aggregatedInstructions := ""
		for _, blueprint := range plotter.Status.Blueprints {
			if len(blueprint.Status.ObservedState.DataAccessInstructions) > 0 {
				aggregatedInstructions = aggregatedInstructions + blueprint.Status.ObservedState.DataAccessInstructions + "\n"
			}
		}
		plotter.Status.ObservedState.DataAccessInstructions = aggregatedInstructions
		// TODO use different RequeueAfter time when plotter is ready?
	}

	// TODO Once a better notification mechanism exists in razee switch to that
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func SetupPlotterController(mgr manager.Manager, clusterManager multicluster.ClusterManager) {
	setupLog := ctrl.Log.WithName("setup")

	if err := NewPlotterReconciler(mgr, "PlotterController", clusterManager).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Plotter")
		os.Exit(1)
	}
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&app.Plotter{}).
		Complete(r)
}
