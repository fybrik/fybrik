// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"
	"encoding/json"
	"path"
	"sort"

	motionv1 "github.com/ibm/the-mesh-for-data/manager/apis/motion/v1alpha1"
	kbatch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Returns the volume configuration for a BatchTransfer.
// This includes the configuration secret as well as
// possible secrets that have to be mounted that contain trust stores
// and additional secrets that are not defined in the configuration secret.
// TODO: Make this usable with the motionv1.Transfer so that it works with different as well
func VolumeConfiguration(batchTransfer *motionv1.BatchTransfer) ([]v1.Volume, []v1.VolumeMount, error) {
	volumes := []v1.Volume{{
		Name: motionv1.ConfigSecretVolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: batchTransfer.Name,
			},
		},
	}, {
		Name: "workspace",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}}
	volumeMounts := []v1.VolumeMount{{
		Name:      motionv1.ConfigSecretVolumeName,
		MountPath: motionv1.ConfigSecretMountPath,
	}, {
		Name:      "workspace",
		MountPath: "/opt/spark/work-dir",
	}}
	if batchTransfer.Spec.Source.Kafka != nil {
		if batchTransfer.Spec.Source.Kafka.SslTruststoreSecret != "" {
			volumes = append(volumes, v1.Volume{
				Name: "source-truststore",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: batchTransfer.Spec.Source.Kafka.SslTruststoreSecret,
					},
				},
			})
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      "source-truststore",
				MountPath: path.Dir(batchTransfer.Spec.Source.Kafka.SslTruststoreLocation),
			})
		}

		if batchTransfer.Spec.Source.Kafka.SecretImport != nil {
			volumes = append(volumes, v1.Volume{
				Name: "source-truststore",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: *(batchTransfer.Spec.Source.Kafka.SecretImport),
					},
				},
			})
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      "source-truststore",
				MountPath: "/etc/secrets/" + *(batchTransfer.Spec.Source.Kafka.SecretImport),
			})
		}
	}
	return volumes, volumeMounts, nil
}

// This is a helper method that returns a kubernetes job description (or an error) that implements the data movement
// described by the given BatchTransfer object. The job is intended to move tabular (relational) data.
// An error is returned if this controller cannot be made the "owner" of this job.
// Ownership determines when the garbage collector in K8s cleans up the job.
func (reconciler *BatchTransferReconciler) createSparkJob(batchTransfer *motionv1.BatchTransfer) (*kbatch.Job, error) {
	// The following piece of code defines a job that implements the actual data movement based on the
	// We want job names for a given nominal start time to have a deterministic name to avoid the same job being
	// created twice
	con := int32(batchTransfer.Spec.MaxFailedRetries)

	volumes, volumeMounts, err := VolumeConfiguration(batchTransfer)

	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations["sidecar.istio.io/inject"] = "false"

	job := &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: annotations,
			Name:        batchTransfer.Name,
			Namespace:   batchTransfer.Namespace,
		},
		Spec: kbatch.JobSpec{
			BackoffLimit: &con,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
				},
				Spec: v1.PodSpec{
					Volumes: volumes,
					Containers: []v1.Container{{
						Name:            "transfer",
						Image:           batchTransfer.Spec.Image,
						ImagePullPolicy: batchTransfer.Spec.ImagePullPolicy,
						Args:            []string{motionv1.ConfigSecretMountPath + "/conf.json"},
						Command:         []string{motionv1.BatchtransferBinary},
						Env: []v1.EnvVar{
							{
								Name: "MY_POD_IP",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "status.podIP",
									},
								},
							},
							{
								Name: "MY_NODE_NAME",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.name",
									},
								},
							},
							{
								Name: "NAMESPACE",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							},
							{
								Name:  "OWNER_NAME",
								Value: batchTransfer.Name,
							},
							{
								Name:  "OWNER_KIND",
								Value: batchTransfer.Kind,
							},
							{
								Name:  "OWNER_UID",
								Value: string(batchTransfer.UID),
							},
						},
						VolumeMounts: volumeMounts,
					}},
					RestartPolicy: "Never",
				},
			},
		},
	}
	if err = ctrl.SetControllerReference(batchTransfer, job, reconciler.Scheme); err != nil {
		return nil, err
	}

	return job, nil
}

// This helper method returns a job that transfers binary data between COS buckets (within the same COS instance).
// TODO: This should be replaced in future. Maybe an image that just does binary copies from COS to COS makes more sense.
func (reconciler *BatchTransferReconciler) createBinaryCopyJob(batchTransfer *motionv1.BatchTransfer) (*kbatch.Job, error) {
	// We want job names for a given nominal start time to have a deterministic name to avoid the same job being
	// created twice
	log := reconciler.Log.WithValues("batchtransfer", batchTransfer.Name)
	log.V(1).Info("Creating binary job!")
	con := int32(1)
	job := &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        batchTransfer.Name,
			Namespace:   batchTransfer.Namespace,
		},
		Spec: kbatch.JobSpec{
			BackoffLimit: &con,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{{
						Name: motionv1.ConfigSecretVolumeName,
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: batchTransfer.Name,
							},
						},
					}},
					Containers: []v1.Container{{
						Name: "transfer",
						//Image: "datamover:1.0-SNAPSHOT",
						Image:           batchTransfer.Spec.Image,
						ImagePullPolicy: batchTransfer.Spec.ImagePullPolicy,
						Args:            []string{motionv1.ConfigSecretMountPath + "/conf.json"},
						Command:         []string{"java", "-cp", "/opt/spark/jars/*:/app/classpath/*:/app/libs/*", "com.ibm.zurich.fabric.mover.BinaryCopy"},
						VolumeMounts: []v1.VolumeMount{{
							Name:      motionv1.ConfigSecretVolumeName,
							MountPath: motionv1.ConfigSecretMountPath,
						}},
					}},
					RestartPolicy: "Never",
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(batchTransfer, job, reconciler.Scheme); err != nil {
		return nil, err
	}

	return job, nil
}

func (reconciler *BatchTransferReconciler) constructBatchJob(batchTransfer *motionv1.BatchTransfer) (*kbatch.Job, error) {
	if batchTransfer.Spec.Source.S3 != nil && batchTransfer.Spec.Source.S3.DataFormat == "binary" {
		return reconciler.createBinaryCopyJob(batchTransfer)
	}
	return reconciler.createSparkJob(batchTransfer)
}

// Create one of two jobs: A spark job for tabular data or a binary copy job.
func (reconciler *BatchTransferReconciler) CreateBatchJob(batchTransfer *motionv1.BatchTransfer) error {
	job, err := reconciler.constructBatchJob(batchTransfer)

	if err != nil {
		return err
	}

	// ...and create it on the cluster
	if err = reconciler.Create(context.Background(), job); err != nil {
		reconciler.Log.Error(err, "unable to create Job for batchTransfer", "job", *job)
		return err
	}

	reconciler.Log.V(1).Info("created Job for batchTransfer run", "job", *job)
	return nil
}

// Create the secret that is used for the BatchTransfer object
// This secret contains the spec of the BatchTransfer object in a JSON file.
func (reconciler *BatchTransferReconciler) CreateSecret(batchTransfer *motionv1.BatchTransfer) error {
	log := reconciler.Log.WithValues("batchtransfer", batchTransfer.Name)

	bytes, err := json.Marshal(batchTransfer.Spec) // Write spec into secret
	if err != nil {
		log.V(1).Info("Error!")
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        batchTransfer.Name,
			Namespace:   batchTransfer.Namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Data: make(map[string][]byte),
	}

	secret.Data["conf.json"] = bytes
	if err := ctrl.SetControllerReference(batchTransfer, secret, reconciler.Scheme); err != nil {
		return err
	}

	if err := reconciler.Create(context.Background(), secret); err != nil {
		log.Error(err, "unable to create Secret for batchTransfer", "secret", secret)
		return err
	}

	log.V(1).Info("created Secret for batchTransfer run", "secret", secret)

	return nil
}

// Returns the condition type of the given job
func isJobFinished(job *kbatch.Job) (bool, kbatch.JobConditionType) {
	// Conditions is an array of a condition and condition is a struct that defines the type of the job condition
	// and the status (True, False, Unknown). Check whether the job is completed or failed by checking each
	// condition in the job's conditions array.
	for _, c := range job.Status.Conditions {
		if (c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed) && c.Status == v1.ConditionTrue {
			return true, c.Type
		}
	}
	// nothing found, so job isn't finished.
	return false, ""
}

// Get the error string from the last pod of the job and set it as error message
// for the BatchTransfer object.
func (reconciler *BatchTransferReconciler) updateBatchErrorMessage(transfer *motionv1.BatchTransfer, controllerID string) {
	log := reconciler.Log.WithValues("batchtransfer", transfer.Name)
	var podList v1.PodList
	ns := client.InNamespace(transfer.Namespace)
	ls := client.MatchingLabels{"controller-uid": controllerID}
	if err := reconciler.List(context.Background(), &podList, ns, ls); err != nil {
		log.Error(err, "unable to list child Pods")
	}

	if len(podList.Items) > 0 {
		sort.SliceStable(podList.Items, func(i, j int) bool {
			return podList.Items[i].Status.StartTime.Time.After(podList.Items[j].Status.StartTime.Time)
		})
		pod := podList.Items[0]
		if len(pod.Status.ContainerStatuses) > 0 {
			message := pod.Status.ContainerStatuses[0].State.Terminated.Message
			transfer.Status.Error = message
		}
	}
}
