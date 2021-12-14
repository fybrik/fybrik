// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	assetmetadata "fybrik.io/fybrik/manager/controllers/app/assetmetadata"
	"fybrik.io/fybrik/pkg/multicluster"
	model "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
)

// WorkloadInfo holds workload details such as the cluster where the workload is running,
// and additional properties defined in the taxonomy, e.g. workload type
type WorkloadInfo struct {
	// Unique fybrikapplication id used for logging
	UUID string `json:"uuid"`
	// Policy set id to allow evaluation of a specific set of policies per fybrikapplication
	PolicySetID string `json:"policySetID"`
	// Cluster where the user workload is running
	Cluster multicluster.Cluster `json:"cluster"`
	// Application/workload properties
	Properties api.ApplicationDetails `json:"properties,omitempty"`
}

// DataRequest is a request to use a specific asset
type DataRequest struct {
	// asset identifier
	DatasetID string `json:"datasetID"`
	// requested interface
	Interface api.InterfaceDetails `json:"interface"`
	// requested usage, e.g. "read": true, "write": false
	Usage map[api.DataFlow]bool `json:"usage"`
	// Asset metadata
	Metadata *assetmetadata.DataDetails `json:"dataset"`
}

// EvaluatorInput is an input to Configuration Policies Evaluator.
// Used to evaluate configuration policies.
type EvaluatorInput struct {
	// Workload configuration
	Workload WorkloadInfo `json:"workload"`
	// Requirements for asset usage
	Request DataRequest `json:"request"`
	// Governance Actions for reading data (relevant for read scenarios only)
	GovernanceActions []model.Action `json:"actions"`
}
