// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"
	"path"

	motionv1 "github.com/ibm/the-mesh-for-data/manager/apis/motion/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"encoding/json"

	apps "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Create the deployment for a stream transfer object.
// This deployment will mount the PVC for the checkpoint directory as well as the secret for the configuration.
func (reconciler *StreamTransferReconciler) CreateDeployment(streamTransfer *motionv1.StreamTransfer) error {
	log := reconciler.Log.WithValues("streamTransfer", streamTransfer.Name)
	replicas := int32(1)
	name := streamTransfer.ObjectMeta.Name
	if streamTransfer.Spec.Suspend {
		replicas = int32(0)
	}
	labels := make(map[string]string)
	labels["streamTransfer"] = name

	annotations := make(map[string]string)
	annotations["sidecar.istio.io/inject"] = "false"
	volumes := []v1.Volume{{
		Name: "checkpoint",
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: name,
			},
		}},
		{
			Name: "configuration",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: name,
				},
			},
		}, {
			Name: "workspace",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}
	volumeMounts := []v1.VolumeMount{{
		Name:      "configuration",
		MountPath: "/etc/mover",
	}, {
		Name:      "checkpoint",
		MountPath: "/tmp/checkpoint",
	}, {
		Name:      "workspace",
		MountPath: "/opt/spark/work-dir",
	}}
	if streamTransfer.Spec.Source.Kafka != nil {
		if streamTransfer.Spec.Source.Kafka.SslTruststoreSecret != "" {
			volumes = append(volumes, v1.Volume{
				Name: "source-truststore",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: streamTransfer.Spec.Source.Kafka.SslTruststoreSecret,
					},
				},
			})
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      "source-truststore",
				MountPath: path.Dir(streamTransfer.Spec.Source.Kafka.SslTruststoreLocation),
			})
		}
	}
	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: streamTransfer.Namespace,
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: v1.PodSpec{
					Volumes: volumes,
					Containers: []v1.Container{{
						Name:            "transfer",
						Image:           streamTransfer.Spec.Image,
						ImagePullPolicy: streamTransfer.Spec.ImagePullPolicy,
						Args:            []string{"/etc/mover/conf.json"},
						Command:         []string{"/streaming-entrypoint.sh"},
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
								Value: streamTransfer.Name,
							},
							{
								Name:  "OWNER_KIND",
								Value: streamTransfer.Kind,
							},
							{
								Name:  "OWNER_UID",
								Value: string(streamTransfer.UID),
							},
						},
						VolumeMounts: volumeMounts,
					}},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(streamTransfer, deployment, reconciler.Scheme); err != nil {
		return err
	}

	if err := reconciler.Create(context.Background(), deployment); err != nil {
		log.Error(err, "unable to create deployment for streamTransfer")
		return err
	}

	log.V(1).Info("created deployment for streamTransfer run")

	return nil
}

// Create the persistent volume claim for the checkpoint directory
// This PVC stores the state of a given transfer so that it can be restarted from a given checkpoint in case of failure.
func (reconciler *StreamTransferReconciler) CreatePVC(streamTransfer *motionv1.StreamTransfer) error {
	log := reconciler.Log.WithValues("streamTransfer", streamTransfer.Name)

	volumeMode := v1.PersistentVolumeFilesystem
	quantity, _ := resource.ParseQuantity("5Gi")
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      streamTransfer.Name,
			Namespace: streamTransfer.Namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			VolumeMode:  &volumeMode,
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{"storage": quantity},
			},
		},
	}

	if err := ctrl.SetControllerReference(streamTransfer, pvc, reconciler.Scheme); err != nil {
		return err
	}

	if err := reconciler.Create(context.Background(), pvc); err != nil {
		log.Error(err, "unable to create PVC for streamTransfer")
		return err
	}

	log.V(1).Info("created PVC for streamTransfer run")

	return nil
}

// Create the secret that is used for the StreamTransfer object
// This secret contains the spec of the StreamTransfer object in a JSON file.
func (reconciler *StreamTransferReconciler) CreateSecret(streamTransfer *motionv1.StreamTransfer) error {
	log := reconciler.Log.WithValues("streamtransfer", streamTransfer.Name)

	bytes, err := json.Marshal(streamTransfer.Spec) // Write spec into secret
	if err != nil {
		log.V(1).Info("Error!")
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        streamTransfer.Name,
			Namespace:   streamTransfer.Namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Data: make(map[string][]byte),
	}

	secret.Data["conf.json"] = bytes
	if err := ctrl.SetControllerReference(streamTransfer, secret, reconciler.Scheme); err != nil {
		return err
	}

	if err := reconciler.Create(context.Background(), secret); err != nil {
		log.Error(err, "unable to create Secret for batchTransfer")
		return err
	}

	log.V(1).Info("created Secret for batchTransfer run")

	return nil
}
