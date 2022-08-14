// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"io"

	"fybrik.io/fybrik/pkg/model/policymanager"
)

// PolicyManager is an interface of a facade to connect to a policy manager.
type PolicyManager interface {
	GetPoliciesDecisions(in *policymanager.GetPolicyDecisionsRequest, creds string) (*policymanager.GetPolicyDecisionsResponse, error)
	io.Closer
}
