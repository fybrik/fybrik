package dummy

import (
	"errors"

	"github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
)

// This ClusterManager is meant to be used for testing
type ClusterManager struct {
	DeployedBlueprints map[string]*v1alpha1.Blueprint
}

func (m *ClusterManager) GetClusters() ([]multicluster.Cluster, error) {
	return []multicluster.Cluster{
		{
			Name:     "kind-kind",
			Metadata: multicluster.ClusterMetadata{},
		},
	}, nil
}

func (m *ClusterManager) GetBlueprint(cluster string, namespace string, name string) (*v1alpha1.Blueprint, error) {
	blueprint, found := m.DeployedBlueprints[cluster]
	if found {
		return blueprint, nil
	}
	return nil, errors.New("Blueprint not found")
}

func (m *ClusterManager) CreateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	m.DeployedBlueprints[cluster] = blueprint
	return nil
}

func (m *ClusterManager) UpdateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	m.DeployedBlueprints[cluster] = blueprint
	return nil
}

func (m *ClusterManager) DeleteBlueprint(cluster string, namespace string, name string) error {
	delete(m.DeployedBlueprints, cluster)
	return nil
}
