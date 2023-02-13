// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policymanager

import (
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

type GetPolicyDecisionsRequest struct {
	Context  taxonomy.PolicyManagerRequestContext `json:"context,omitempty"`
	Action   RequestAction                        `json:"action"`
	Resource Resource                             `json:"resource"`
}

type GetPolicyDecisionsResponse struct {
	DecisionID string `json:"decision_id,omitempty"`
	// Additional message to be reported to the user
	Message string `json:"message,omitempty"`
	// Result of policy evaluation
	Result []ResultItem `json:"result"`
}
