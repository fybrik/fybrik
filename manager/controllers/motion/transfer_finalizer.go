// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"

	motionv1 "github.com/ibm/the-mesh-for-data/manager/apis/motion/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Add a finalizer to the given transfer and update it.
func (reconciler *Reconciler) addFinalizer(transfer motionv1.Transfer) error {
	transfer.AddFinalizer()
	return reconciler.Update(context.Background(), transfer)
}

// Remove a finalizer from the given transfer and update it.
func (reconciler *Reconciler) removeFinalizer(transfer motionv1.Transfer) error {
	transfer.RemoveFinalizer()
	return reconciler.Update(context.Background(), transfer)
}

// Submits a finalizer pod that calls the finalizer entrypoint of a transfer image.
func (reconciler *Reconciler) submitFinalizerPod(transfer motionv1.Transfer) error {
	reconciler.Log.Info("Creating finalizer pod...")
	// Finalizer Pod does not exist. Create it

	annotations := make(map[string]string)
	annotations["sidecar.istio.io/inject"] = "false"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        transfer.FinalizerPodName(),
			Namespace:   transfer.GetNamespace(),
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: motionv1.ConfigSecretVolumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: transfer.GetName(),
					},
				},
			}},
			Containers: []corev1.Container{{
				Name:            "transfer",
				Image:           transfer.GetImage(),
				ImagePullPolicy: transfer.GetImagePullPolicy(),
				Command:         []string{motionv1.BatchtransferFinalizerBinary},
				Args:            []string{motionv1.ConfigSecretMountPath + "/conf.json"},
				VolumeMounts: []corev1.VolumeMount{{
					Name:      motionv1.ConfigSecretVolumeName,
					MountPath: motionv1.ConfigSecretMountPath,
				}},
			}},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	if err := ctrl.SetControllerReference(transfer, pod, reconciler.Scheme); err != nil {
		return err
	}

	if err := reconciler.Create(context.Background(), pod); err != nil {
		reconciler.Log.Error(err, "unable to create finalizer process for transfer", "pod", pod)
		return err
	}

	return nil
}

// Handle the finalizer logic of a transfer.
// The finalizer will spawn a pod that will execute the datamover with a special endpoint that knows how to remove
// the target for each datastore type.
func (reconciler *Reconciler) handleFinalizer(transfer motionv1.Transfer) error {
	if !transfer.HasFinalizer() {
		return nil
	}

	if !transfer.HasStarted() {
		// If there are no last jobs there also have not been any transfers. Nothing to tidy up -> remove finalizer
		if err := reconciler.removeFinalizer(transfer); err != nil {
			return err
		}
	}

	// Handle actual finalizer logic

	var existingPod corev1.Pod
	if err := reconciler.Get(context.Background(), transfer.FinalizerPodKey(), &existingPod); err != nil {
		if kerrors.IsNotFound(err) {
			// Finalizer pod is not found. Create it
			if err := reconciler.submitFinalizerPod(transfer); err != nil {
				return err
			}
			return nil // After the finalizer job has been created just return
		}
	}

	// Handle Finalizer pod run
	switch {
	case existingPod.Status.Phase == corev1.PodSucceeded:
		reconciler.Log.Info("Finalizer finished successfully")
		// Pod succeeded. Remove finalizer and let crd be deleted
		if err := reconciler.removeFinalizer(transfer); err != nil {
			return err
		}
	case existingPod.Status.Phase == corev1.PodFailed:
		reconciler.Log.Info("Finalizer pod failed!!")
		// Pod failed. Remove finalizer and let crd be deleted as blocking is also not an option.
		if err := reconciler.removeFinalizer(transfer); err != nil {
			return err
		}
	}
	return nil
}
