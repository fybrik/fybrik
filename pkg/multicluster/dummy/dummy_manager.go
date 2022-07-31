// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package dummy

import (
	"errors"
	app "fybrik.io/fybrik/manager/apis/app/v1"
	"fybrik.io/fybrik/pkg/multicluster"
)

// MockClusterManager is meant to be used for testing
type MockClusterManager struct {
	DeployedBlueprints map[string]*app.Blueprint
	Clusters           []multicluster.Cluster
}

func NewDummyClusterManager(blueprints map[string]*app.Blueprint, clusters []multicluster.Cluster) MockClusterManager {
	return MockClusterManager{
		DeployedBlueprints: blueprints,
		Clusters:           clusters,
	}
}

func (m *MockClusterManager) GetClusters() ([]multicluster.Cluster, error) {
	if m.Clusters != nil {
		return m.Clusters, nil
	}
	return []multicluster.Cluster{
		{
			Name:     "kind-kind",
			Metadata: multicluster.ClusterMetadata{},
		},
	}, nil
}

func (m *MockClusterManager) IsMultiClusterSetup() bool {
	return true
}

func (m *MockClusterManager) GetBlueprint(cluster, namespace, name string) (*app.Blueprint, error) {
	blueprint, found := m.DeployedBlueprints[cluster]
	if found {
		return blueprint, nil
	}
	return nil, errors.New("blueprint not found")
}

func (m *MockClusterManager) CreateBlueprint(cluster string, blueprint *app.Blueprint) error {
	m.DeployedBlueprints[cluster] = blueprint
	return nil
}

func (m *MockClusterManager) UpdateBlueprint(cluster string, blueprint *app.Blueprint) error {
	m.DeployedBlueprints[cluster] = blueprint
	return nil
}

func (m *MockClusterManager) DeleteBlueprint(cluster, namespace, name string) error {
	delete(m.DeployedBlueprints, cluster)
	return nil
}
