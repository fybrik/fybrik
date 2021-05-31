// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	motionv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/motion/v1alpha1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// A reconciler can be used to base reconcilers of Transfers on.
// It has functions on how to manage finalizers for transfers
// and other utilities that can be used by specific reconcilers.
// It is "derived" (using type embedding) from the K8s client
type Reconciler struct {
	client.Client
	Name   string
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// BatchTransferReconciler reconciles a BatchTransfer object
// It is "derived" from the Reconciler object
type BatchTransferReconciler struct {
	Reconciler
}

// This is the main entry point of the controller. It reconciles BatchTransfer objects.
// The batch transfer is implemented as a K8s Job or CronJob.
// Reconciliation happens with the following steps:
// - Fetch the BatchTransfer object
// - Check if the object is being deleted and handle a finalizer if needed
// - Update the status by checking the existing Job/CronJob
// - If K8s objects are not yet created create the objects
func (reconciler *BatchTransferReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := reconciler.Log.WithValues("batchtransfer", req.NamespacedName)

	batchTransfer := &motionv1.BatchTransfer{}
	if err := reconciler.Get(ctx, req.NamespacedName, batchTransfer); err != nil {
		if err.(*kerrors.StatusError).ErrStatus.Code != 404 {
			log.Error(err, "unable to fetch BatchTransfer")
		}
		// ignore not-found errors since they can't be fixed by an immediate requeue.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle the finalizer if the object is being deleted
	if batchTransfer.IsBeingDeleted() {
		if err := reconciler.handleFinalizer(batchTransfer); err != nil {
			log.Error(err, "error when handling the finalizer")
			return ctrl.Result{}, err
		}
	}

	// Add the finalizer if needed
	if !batchTransfer.HasFinalizer() && !batchTransfer.Spec.NoFinalizer {
		if err := reconciler.addFinalizer(batchTransfer); err != nil {
			log.Info(fmt.Sprintf("Could not register finalizer for %s", batchTransfer.Name))
			return ctrl.Result{}, err
		}
	}

	// Reconcile status
	if !batchTransfer.IsCronJob() {
		job := &kbatch.Job{}
		if err := reconciler.Get(ctx, batchTransfer.ObjectKey(), job); err != nil {
			if !kerrors.IsNotFound(err) {
				log.Error(err, "could not fetch Job for batchTransfer %s", "batchTransfer", batchTransfer.ObjectKey())
				return ctrl.Result{}, err
			}
			// If job was not found just continue...
		} else {
			_, finishedType := isJobFinished(job)
			jobRef, err := reference.GetReference(reconciler.Scheme, job)
			if err != nil {
				log.Error(err, "unable to make reference to active job", "job", job)
			}

			// update the list of jobs by state.
			switch finishedType {
			case "": // not a finisher, i.e. still on-going.
				if job.Status.Active == 0 {
					batchTransfer.Status.Status = motionv1.Starting
				} else {
					batchTransfer.Status.Status = motionv1.Running
				}
				batchTransfer.Status.Active = jobRef
			case kbatch.JobFailed:
				batchTransfer.Status.Status = motionv1.Failed
				batchTransfer.Status.LastFailed = jobRef
				batchTransfer.Status.Active = nil
				reconciler.updateBatchErrorMessage(batchTransfer, job.Labels["controller-uid"])
			case kbatch.JobComplete:
				batchTransfer.Status.Status = motionv1.Succeeded
				batchTransfer.Status.LastCompleted = jobRef
				batchTransfer.Status.Active = nil
			}

			// update the status of our CRD.
			if err := reconciler.Status().Update(ctx, batchTransfer); err != nil {
				log.Error(err, "unable to update batchTransfer status")
				return ctrl.Result{}, err
			}
		}
	} else {
		// TODO Handle CronJob status updates
		return ctrl.Result{}, nil
	}

	// Make sure that the secret exists
	existingSecret := &corev1.Secret{}
	if err := reconciler.Get(ctx, batchTransfer.ObjectKey(), existingSecret); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "could not fetch Secret for batchTransfer %s", batchTransfer.ObjectKey())
			return ctrl.Result{}, err
		}

		if err := reconciler.CreateSecret(batchTransfer); err != nil {
			if kerrors.IsAlreadyExists(err) {
				log.Info(fmt.Sprintf("Secret %s already exists!", batchTransfer.Name))
			} else {
				log.Error(err, "unable to create secret!")
				return ctrl.Result{}, err
			}
		}
	}

	// Create new jobs if job has not yet started
	if !batchTransfer.HasStarted() {
		if !batchTransfer.IsCronJob() {
			// Start normal batch job

			if err := reconciler.CreateBatchJob(batchTransfer); err != nil {
				if kerrors.IsAlreadyExists(err) {
					log.Info(fmt.Sprintf("Job %s already exists!", batchTransfer.Name))
				} else {
					return ctrl.Result{}, err
				}
			}
		} else {
			// If batch job is suspended don't do anything
			if batchTransfer.Spec.Suspend {
				log.V(1).Info("batchjob suspended, skipping")
				return ctrl.Result{}, nil
			}
			if err := reconciler.createCronJob(batchTransfer); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// Setup the reconciler. This consists of creating an index of jobs where this controller is the owner.
func (reconciler *BatchTransferReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&motionv1.BatchTransfer{}).
		Owns(&kbatch.Job{}).
		Owns(&corev1.Pod{}).
		Complete(reconciler)
}

// NewBatchTransferReconciler creates a new reconciler for BatchTransfer resources
func NewBatchTransferReconciler(mgr ctrl.Manager, name string) *BatchTransferReconciler {
	return &BatchTransferReconciler{
		Reconciler{
			Client: mgr.GetClient(),
			Name:   name,
			Log:    ctrl.Log.WithName("controllers").WithName(name),
			Scheme: mgr.GetScheme(),
		},
	}
}
