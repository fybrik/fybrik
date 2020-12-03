package local

import (
	"context"
	"github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	m4dSystemNs                  string = "m4d-system"
	clusterMetadataConfigmapName string = "cluster-metadata"
)

type LocalClusterManager struct {
	Client client.Client
}

func (cm *LocalClusterManager) GetClusters() ([]multicluster.Cluster, error) {
	clusterMetadataConfigmap := corev1.ConfigMap{}
	namespacedName := client.ObjectKey{
		Name:      clusterMetadataConfigmapName,
		Namespace: m4dSystemNs,
	}
	if err := cm.Client.Get(context.Background(), namespacedName, &clusterMetadataConfigmap); err != nil {
		return nil, err
	}
	var clusters []multicluster.Cluster
	cluster := multicluster.Cluster{
		Name: clusterMetadataConfigmap.Data["ClusterName"],
		Metadata: multicluster.ClusterMetadata{
			Region: clusterMetadataConfigmap.Data["Region"],
			Zone:   clusterMetadataConfigmap.Data["Zone"],
		},
	}
	clusters = append(clusters, cluster)
	return clusters, nil
}

func (cm *LocalClusterManager) GetBlueprint(cluster string, namespace string, name string) (*v1alpha1.Blueprint, error) {
	return nil, nil
}

func (cm *LocalClusterManager) CreateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	return nil
}

func (cm *LocalClusterManager) UpdateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	return nil
}

func (cm *LocalClusterManager) DeleteBlueprint(cluster string, namespace string, name string) error {
	return nil
}

func CreateLocalClusterManager(client client.Client) multicluster.ClusterManager {
	return &LocalClusterManager{
		Client: client,
	}
}
