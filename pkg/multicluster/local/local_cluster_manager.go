// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package local

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/multicluster"
)

const (
	clusterMetadataConfigmapName string = "cluster-metadata"
)

// localClusterManager for local cluster configuration
type localClusterManager struct {
	Client    client.Client
	Namespace string
}

// GetClusters returns a list of registered clusters
func (cm *localClusterManager) GetClusters() ([]multicluster.Cluster, error) {
	cmcm := corev1.ConfigMap{}
	namespacedName := client.ObjectKey{
		Name:      clusterMetadataConfigmapName,
		Namespace: cm.Namespace,
	}
	if err := cm.Client.Get(context.Background(), namespacedName, &cmcm); err != nil {
		return nil, errors.Wrap(err, "error in GetClusters")
	}
	cluster := multicluster.CreateCluster(cmcm)
	clusters := []multicluster.Cluster{cluster}
	return clusters, nil
}

// GetLocalClusterName returns the local cluster name
func (cm *localClusterManager) GetLocalClusterName() (string, error) {
	clusters, err := cm.GetClusters()
	if err != nil {
		return "", err
	}
	if len(clusters) != 1 {
		return "", errors.New(v1alpha1.InvalidClusterConfiguration)
	}
	return clusters[0].Name, nil
}

// GetBlueprint returns a blueprint matching the given name, namespace and cluster details
func (cm *localClusterManager) GetBlueprint(cluster, namespace, name string) (*v1alpha1.Blueprint, error) {
	if localCluster, err := cm.GetLocalClusterName(); err != nil || localCluster != cluster {
		return nil, fmt.Errorf("unregistered cluster: %s", cluster)
	}
	blueprint := &v1alpha1.Blueprint{}
	namespacedName := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	err := cm.Client.Get(context.Background(), namespacedName, blueprint)
	return blueprint, err
}

// CreateBlueprint creates a blueprint resource or updates an existing one
func (cm *localClusterManager) CreateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	return cm.UpdateBlueprint(cluster, blueprint)
}

// UpdateBlueprint updates the given blueprint or creates a new one if it does not exist
func (cm *localClusterManager) UpdateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	if localCluster, err := cm.GetLocalClusterName(); err != nil || localCluster != cluster {
		return fmt.Errorf("unregistered cluster: %s", cluster) //nolint:revive // Ignore repetitive error msg
	}
	resource := &v1alpha1.Blueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      blueprint.Name,
			Namespace: blueprint.Namespace,
		},
	}
	if _, err := ctrl.CreateOrUpdate(context.Background(), cm.Client, resource, func() error {
		resource.Spec = blueprint.Spec
		resource.ObjectMeta.Finalizers = blueprint.ObjectMeta.Finalizers
		resource.ObjectMeta.Labels = blueprint.ObjectMeta.Labels
		resource.ObjectMeta.Annotations = blueprint.ObjectMeta.Annotations
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// DeleteBlueprint deletes the blueprint resource
func (cm *localClusterManager) DeleteBlueprint(cluster, namespace, name string) error {
	blueprint, err := cm.GetBlueprint(cluster, namespace, name)
	if err != nil {
		return err
	}
	return cm.Client.Delete(context.Background(), blueprint)
}

// NewClusterManager creates an instance of ClusterManager for a local cluster configuration
func NewClusterManager(cl client.Client, namespace string) (multicluster.ClusterManager, error) {
	return &localClusterManager{
		Client:    cl,
		Namespace: namespace,
	}, nil
}
