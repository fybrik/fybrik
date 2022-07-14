// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"fybrik.io/fybrik/pkg/multicluster"
)

const (
	VaultAuthPath = "kubernetes"
)

// ClusterLister is a mockup cluster manager
type ClusterLister struct {
}

func (m *ClusterLister) IsMultiClusterSetup() bool {
	return true
}

// GetClusters returns the cluster config for testing
func (m *ClusterLister) GetClusters() ([]multicluster.Cluster, error) {
	return []multicluster.Cluster{
		{
			Name:     "thegreendragon",
			Metadata: multicluster.ClusterMetadata{Region: "theshire", VaultAuthPath: VaultAuthPath},
		},
		{
			Name:     "neverland-cluster",
			Metadata: multicluster.ClusterMetadata{Region: "neverland", VaultAuthPath: VaultAuthPath},
		},
	}, nil
}
