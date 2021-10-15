// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package evaluator

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	dataset "fybrik.io/fybrik/manager/controllers/app/dataset"
	"fybrik.io/fybrik/pkg/multicluster"
	model "fybrik.io/fybrik/pkg/taxonomy/model/base"
)

// WorkloadInfo holds workload details such as cluster/region, type, etc.
type WorkloadInfo struct {
	// Cluster where the user workload is running
	Cluster multicluster.Cluster
}

// Request is a request to use a specific asset
type Request struct {
	// asset identifier
	DatasetID string
	// requested interface
	Interface api.InterfaceDetails
	// requested usage, e.g. "read": true, "write": false
	Usage map[api.DataFlow]bool
}

// EvaluatorInput is an input to ConfigurationPoliciesEvaluator.
// Used to evaluate configuration policies.
type EvaluatorInput struct {
	// A list of available clusters
	Clusters []multicluster.Cluster
	// Workload configuration
	Workload WorkloadInfo
	// Application properties
	AppInfo api.ApplicationDetails
	// Asset metadata
	AssetMetadata *dataset.DataDetails
	// Requirements for asset usage
	AssetRequirements Request
	// Governance Actions for reading data
	GovernanceActions []model.Action
}

// SetWorkloadInfo generates a new WorkloadInfo object based on FybrikApplication and updates the input structure
// If no cluster has been specified for a workload, a local cluster is assumed.
func SetWorkloadInfo(localCluster multicluster.Cluster, clusters []multicluster.Cluster, application *api.FybrikApplication, result *EvaluatorInput) {
	clusterName := application.Spec.Selector.ClusterName
	if clusterName == "" {
		result.Workload.Cluster = localCluster
	} else {
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				result.Workload.Cluster = localCluster
				break
			}
		}
	}
}

// SetApplicationInfo generates a new AppInfo object based on FybrikApplication and updates the input structure
func SetApplicationInfo(application *api.FybrikApplication, result *EvaluatorInput) {
	result.AppInfo = application.Spec.AppInfo.DeepCopy()
}

// SetAssetRequirements generates a new Request object for a specific asset based on FybrikApplication and updates the input structure
func SetAssetRequirements(application *api.FybrikApplication, dataCtx api.DataContext, result *EvaluatorInput) {
	usage := make(map[api.DataFlow]bool)
	usage[api.ReadFlow] = (application.Spec.Selector.WorkloadSelector.Size() > 0)
	usage[api.CopyFlow] = dataCtx.Requirements.Copy.Required
	result.AssetRequirements = Request{
		DatasetID: dataCtx.DataSetID,
		Interface: dataCtx.Requirements.Interface,
		Usage:     usage,
	}
}
