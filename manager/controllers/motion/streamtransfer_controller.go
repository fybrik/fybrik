// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"
	"fmt"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	motionv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/motion/v1alpha1"
)

// StreamTransferReconciler reconciles a StreamTransfer object
// It is "derived" from the Reconciler object
type StreamTransferReconciler struct {
	Reconciler
}

// Reconcile StreamTransfers
// A first version of this StreamTransfer is based on K8s Deployment objects.
// These manage pod failures themselves and restart them in case of failure.
// Thus in this first version that does not handle errors yet a stream will be started
// if it does not exist (including a persistent checkpoint storage) and otherwise be left running.
// A more involved version that is handling errors and is discovering crash loops may have to be based
// on a Pod directly in order to discover errors on a finer granular level.
func (reconciler *StreamTransferReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := reconciler.Log.WithValues("streamtransfer", req.NamespacedName)

	streamTransfer := &motionv1.StreamTransfer{}
	if err := reconciler.Get(ctx, req.NamespacedName, streamTransfer); err != nil {
		if err.(*kerrors.StatusError).ErrStatus.Code != 404 {
			log.Error(err, "Unable to fetch StreamTransfer")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle the finalizer if the object is being deleted
	if streamTransfer.IsBeingDeleted() {
		if err := reconciler.handleFinalizer(streamTransfer); err != nil {
			log.Error(err, "error when handling the finalizer")
			return ctrl.Result{}, nil
		}
	}

	// Add the finalizer if needed
	if !streamTransfer.HasFinalizer() && !streamTransfer.Spec.NoFinalizer {
		if err := reconciler.addFinalizer(streamTransfer); err != nil {
			log.Info(fmt.Sprintf("Could not register finalizer for %s", streamTransfer.Name))
			return ctrl.Result{}, err
		}
	}

	// Reconcile status
	deployment := &apps.Deployment{}
	if err := reconciler.Get(ctx, streamTransfer.ObjectKey(), deployment); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("could not fetch deployment for StreamTransfer %s", streamTransfer.Name))
			return ctrl.Result{}, nil
		}
	} else {
		// Update state of the StreamJob depending on if the deployment is running or not.
		if deployment.Status.AvailableReplicas == *(deployment.Spec.Replicas) {
			streamTransfer.Status.Status = motionv1.StreamRunning
		} else {
			streamTransfer.Status.Status = motionv1.StreamStarting
		}
		// TODO more granular status for a failing/stopped stream

		if err := reconciler.Status().Update(ctx, streamTransfer); err != nil {
			log.Error(err, "unable to update streamTransfer status")
			return ctrl.Result{}, err
		}
	}

	// Make sure that the secret exists
	existingSecret := &v1.Secret{}
	if err := reconciler.Get(ctx, streamTransfer.ObjectKey(), existingSecret); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "could not fetch Secret for streamTransfer %s", streamTransfer.ObjectKey())
			return ctrl.Result{}, err
		}

		if err := reconciler.CreateSecret(streamTransfer); err != nil {
			if kerrors.IsAlreadyExists(err) {
				log.Info(fmt.Sprintf("Secret %s already exists!", streamTransfer.Name))
			} else {
				log.Error(err, "unable to create secret!")
				return ctrl.Result{}, err
			}
		}
	}

	// Make sure that pvc exists
	existingpvc := &v1.PersistentVolumeClaim{}
	if err := reconciler.Get(ctx, streamTransfer.ObjectKey(), existingpvc); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "could not fetch PVC for streamTransfer %s", streamTransfer.ObjectKey())
			return ctrl.Result{}, err
		}

		if err := reconciler.CreatePVC(streamTransfer); err != nil {
			if kerrors.IsAlreadyExists(err) {
				log.Info(fmt.Sprintf("PVC %s already exists!", streamTransfer.Name))
			} else {
				log.Error(err, "unable to create PVC!")
				return ctrl.Result{}, err
			}
		}
	}

	// Make sure that deployment exists
	if !streamTransfer.HasStarted() {
		if err := reconciler.CreateDeployment(streamTransfer); err != nil {
			if kerrors.IsAlreadyExists(err) {
				log.Info(fmt.Sprintf("PVC %s already exists!", streamTransfer.Name))
			} else {
				log.Error(err, "unable to create PVC!")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (reconciler *StreamTransferReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&motionv1.StreamTransfer{}).
		Owns(&apps.Deployment{}).
		Complete(reconciler)
}

// NewStreamTransferReconciler creates a new reconciler for StreamTransfer resources
func NewStreamTransferReconciler(mgr ctrl.Manager, name string) *StreamTransferReconciler {
	return &StreamTransferReconciler{
		Reconciler{
			Client: mgr.GetClient(),
			Name:   name,
			Log:    ctrl.Log.WithName("controllers").WithName(name),
			Scheme: mgr.GetScheme(),
		},
	}
}
