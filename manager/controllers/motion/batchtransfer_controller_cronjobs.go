// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"

	motionv1 "github.com/ibm/the-mesh-for-data/manager/apis/motion/v1alpha1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Creates a Kubernetes CronJob from a BatchTransfer
// This is used if the schedule field in a BatchTransfer is not empty and a
// BatchTransfer should be scheduled on a regular basis.
// This CronJob requires a secret with the same name for the configuration parameters.
// The specifics of the job template is retrieved from the constructBatchJob method.
// This function directly creates the cron job and returns an error if something went wrong.
func (reconciler *BatchTransferReconciler) createCronJob(batchTransfer *motionv1.BatchTransfer) error {
	successfulJobHistoryLimit := int32(batchTransfer.Spec.SuccessfulJobHistoryLimit)
	failedJobHistoryLimit := int32(batchTransfer.Spec.FailedJobHistoryLimit)

	job, err := reconciler.constructBatchJob(batchTransfer)

	if err != nil {
		return err
	}

	cronJob := &v1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      batchTransfer.Name,
			Namespace: batchTransfer.Namespace,
		},
		Spec: v1beta1.CronJobSpec{
			Schedule: batchTransfer.Spec.Schedule,
			Suspend:  &batchTransfer.Spec.Suspend,
			JobTemplate: v1beta1.JobTemplateSpec{
				Spec: job.Spec,
			},
			SuccessfulJobsHistoryLimit: &successfulJobHistoryLimit,
			FailedJobsHistoryLimit:     &failedJobHistoryLimit,
		},
	}

	if err := ctrl.SetControllerReference(batchTransfer, cronJob, reconciler.Scheme); err != nil {
		return err
	}

	// ...and create it on the cluster
	if err := reconciler.Create(context.Background(), cronJob); err != nil {
		reconciler.Log.Error(err, "unable to create Job for batchTransfer", "job", *cronJob)
		return err
	}

	reconciler.Log.V(1).Info("created Job for batchTransfer run", "job", *cronJob)
	return nil
}
