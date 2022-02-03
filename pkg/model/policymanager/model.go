// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policymanager

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

type RequestAction struct {
	ActionType         api.DataFlow                `json:"actionType"`
	ProcessingLocation taxonomy.ProcessingLocation `json:"processingLocation,omitempty"`
	Destination        string                      `json:"destination,omitempty"`
}

type Resource struct {
	ID       taxonomy.AssetID              `json:"id"`
	Metadata *datacatalog.ResourceMetadata `json:"metadata,omitempty"`
}

type ResultItem struct {
	// The policy on which the decision was based
	Policy string          `json:"policy"`
	Action taxonomy.Action `json:"action"`
}
