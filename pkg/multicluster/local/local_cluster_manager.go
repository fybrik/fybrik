package local

import (
	"context"
	"errors"
	"fmt"

	"github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	clusterMetadataConfigmapName string = "cluster-metadata"
)

// ClusterManager for local cluster configuration
type ClusterManager struct {
	Client    client.Client
	Namespace string
}

// GetClusters returns a list of registered clusters
func (cm *ClusterManager) GetClusters() ([]multicluster.Cluster, error) {
	clusterMetadataConfigmap := corev1.ConfigMap{}
	namespacedName := client.ObjectKey{
		Name:      clusterMetadataConfigmapName,
		Namespace: cm.Namespace,
	}
	if err := cm.Client.Get(context.Background(), namespacedName, &clusterMetadataConfigmap); err != nil {
		wrappedError := fmt.Errorf("Error in GetClusters: %w", err)
		return nil, wrappedError
	}
	var clusters []multicluster.Cluster
	cluster := multicluster.Cluster{
		Name: clusterMetadataConfigmap.Data["ClusterName"],
		Metadata: multicluster.ClusterMetadata{
			Region:        clusterMetadataConfigmap.Data["Region"],
			Zone:          clusterMetadataConfigmap.Data["Zone"],
			VaultAuthPath: clusterMetadataConfigmap.Data["VaultAuthPath"],
		},
	}
	clusters = append(clusters, cluster)
	return clusters, nil
}

// GetLocalClusterName returns the local cluster name
func (cm *ClusterManager) GetLocalClusterName() (string, error) {
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
func (cm *ClusterManager) GetBlueprint(cluster string, namespace string, name string) (*v1alpha1.Blueprint, error) {
	if localCluster, err := cm.GetLocalClusterName(); err != nil || localCluster != cluster {
		return nil, errors.New("Unregistered cluster: " + cluster)
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
func (cm *ClusterManager) CreateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	return cm.UpdateBlueprint(cluster, blueprint)
}

// UpdateBlueprint updates the given blueprint or creates a new one if does not exist
func (cm *ClusterManager) UpdateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	if localCluster, err := cm.GetLocalClusterName(); err != nil || localCluster != cluster {
		return errors.New("Unregistered cluster: " + cluster)
	}
	resource := &v1alpha1.Blueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      blueprint.Name,
			Namespace: blueprint.Namespace,
		},
	}
	if _, err := ctrl.CreateOrUpdate(context.Background(), cm.Client, resource, func() error {
		resource.Spec = blueprint.Spec
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// DeleteBlueprint deletes the blueprint resource
func (cm *ClusterManager) DeleteBlueprint(cluster string, namespace string, name string) error {
	blueprint, err := cm.GetBlueprint(cluster, namespace, name)
	if err != nil {
		return err
	}
	return cm.Client.Delete(context.Background(), blueprint)
}

// NewManager creates a new ClusterManager for a local cluster configuration
func NewManager(client client.Client, namespace string) (multicluster.ClusterManager, error) {
	return &ClusterManager{
		Client:    client,
		Namespace: namespace,
	}, nil
}
