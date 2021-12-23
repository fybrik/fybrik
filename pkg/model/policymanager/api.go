// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policymanager

import (
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// +fybrik:validation:object
type GetPolicyDecisionsRequest struct {
	Context  taxonomy.PolicyManagerRequestContext `json:"context,omitempty"`
	Action   RequestAction                        `json:"action"`
	Resource datacatalog.ResourceMetadata         `json:"resource"`
}

// +fybrik:validation:object
type GetPolicyDecisionsResponse struct {
	DecisionID string       `json:"decision_id,omitempty"`
	Result     []ResultItem `json:"result"`
}
