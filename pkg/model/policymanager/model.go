// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policymanager

import (
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// RequestAction describes the reason for accessing the data, e.g., read/write/delete, where the data is processed or written to
type RequestAction struct {
	ActionType         taxonomy.DataFlow           `json:"actionType"`
	ProcessingLocation taxonomy.ProcessingLocation `json:"processingLocation,omitempty"`
	Destination        string                      `json:"destination,omitempty"`
}

// Asset metadata
type Resource struct {
	ID       taxonomy.AssetID              `json:"id"`
	Metadata *datacatalog.ResourceMetadata `json:"metadata,omitempty"`
}

// Result of policy evaluation
type ResultItem struct {
	// The policy on which the decision was based
	Policy string          `json:"policy"`
	Action taxonomy.Action `json:"action"`
}
