// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"fybrik.io/fybrik/pkg/multicluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Infrastructure details, such as available clusters, storage accounts, metrics.
// TODO(shlomitk1): include available storage accounts
// Metrics (clusters, networking) are not supported yet.
// TODO(shlomitk1): define infrastructure taxonomy to be used in this structure
type Infrastructure struct {
	// Clusters available for deployment
	Clusters []multicluster.Cluster `json:"clusters"`
}

// InfrastructureManager retrieves the infrastructure data, such as ClusterManager interface, kubernetes client, etc.
type InfrastructureManager struct {
	ClusterManager multicluster.ClusterLister
	Client         client.Client
}

// SetInfrastructure uses available interfaces to get the infrastructure details
func (r *InfrastructureManager) SetInfrastructure() (*Infrastructure, error) {
	clusters, err := r.ClusterManager.GetClusters()
	if err != nil {
		return nil, err
	}
	return &Infrastructure{Clusters: clusters}, nil
}
