// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	"fybrik.io/fybrik/pkg/multicluster"
	model "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
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

// EvaluatorInput is an input to Configuration Policies Evaluator.
// Used to evaluate configuration policies.
type EvaluatorInput struct {
	// Workload configuration
	Workload WorkloadInfo
	// Application properties
	AppInfo api.ApplicationDetails
	// Asset metadata
	AssetMetadata api.AssetMetadata
	// Requirements for asset usage
	AssetRequirements Request
	// Governance Actions for reading data (relevant for read scenarios only)
	GovernanceActions []model.Action
}

// SetApplicationInfo generates a new AppInfo object based on FybrikApplication and updates the input structure
func SetApplicationInfo(application *api.FybrikApplication, result *EvaluatorInput) {
	result.AppInfo = application.Spec.AppInfo.DeepCopy()
}

// SetAssetRequirements generates a new Request object for a specific asset based on FybrikApplication and updates the input structure
func SetAssetRequirements(application *api.FybrikApplication, dataCtx api.DataContext, result *EvaluatorInput) {
	usage := make(map[api.DataFlow]bool)
	// request to read is determined by the workload selector presence
	usage[api.ReadFlow] = (application.Spec.Selector.WorkloadSelector.Size() > 0)
	// explicit request to copy
	usage[api.CopyFlow] = dataCtx.Requirements.Copy.Required
	result.AssetRequirements = Request{
		DatasetID: dataCtx.DataSetID,
		Interface: dataCtx.Requirements.Interface,
		Usage:     usage,
	}
}
