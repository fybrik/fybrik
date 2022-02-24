// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
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
	Properties taxonomy.AppInfo `json:"properties,omitempty"`
}

// DataRequest is a request to use a specific asset
type DataRequest struct {
	// asset identifier
	DatasetID string `json:"datasetID"`
	// requested interface
	Interface taxonomy.Interface `json:"interface"`
	// requested usage, e.g. "read": true, "write": false
	Usage taxonomy.DataFlow `json:"usage"`
	// Asset metadata
	Metadata *datacatalog.ResourceMetadata `json:"dataset"`
}

// EvaluatorInput is an input to Configuration Policies Evaluator.
// Used to evaluate configuration policies.
type EvaluatorInput struct {
	// Workload configuration
	Workload WorkloadInfo `json:"workload"`
	// Requirements for asset usage
	Request DataRequest `json:"request"`
	// Governance Actions for reading data (relevant for read scenarios only)
	GovernanceActions []taxonomy.Action `json:"actions"`
}
