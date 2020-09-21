// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policycompilerbl

import (
	"testing"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	tu "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/testutil"
)

func policyManagerMainOnly(purpose string) *pb.PoliciesDecisions {
	mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs, _, _ := tu.GetEnvironment()
	applicationContext := tu.GetApplicationContext(purpose)

	mainPolicyManager := NewPolicyManagerConnector(mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs)
	policiesDecision, _ := mainPolicyManager.GetEnforcementActions(applicationContext)
	return policiesDecision
}

func policyManagerExtOnly(purpose string) *pb.PoliciesDecisions {
	_, _, timeOutSecs, extensionPolicyManagerName, extensionPolicyManagerURL := tu.GetEnvironment()
	applicationContext := tu.GetApplicationContext(purpose)

	extPolicyManager := NewPolicyManagerConnector(extensionPolicyManagerName, extensionPolicyManagerURL, timeOutSecs)
	policiesDecision, _ := extPolicyManager.GetEnforcementActions(applicationContext)
	return policiesDecision
}

//Tests GetEnforcementActions func in policy_manager.go
//tests the main policy manager configuration
//test for purpose "fraud-detection", connector mocks configured for specific output per purpose
func TestMainPolicyManagerPurpose1(t *testing.T) {
	policyDecision := policyManagerMainOnly("fraud-detection")
	mainPMpolicies := tu.GetMainPMDecisions("fraud-detection")
	tu.EnsureDeepEqualDecisions(t, policyDecision, mainPMpolicies)
}

//Tests GetEnforcementActions func in policy_manager.go
//tests the main policy manager configuration
//test for purpose "marketing", connector mocks configured for specific output per purpose
func TestMainPolicyManagerPurpose2(t *testing.T) {
	policyDecision := policyManagerMainOnly("marketing")
	mainPMpolicies := tu.GetMainPMDecisions("marketing")
	tu.EnsureDeepEqualDecisions(t, policyDecision, mainPMpolicies)
}

//Tests GetEnforcementActions func in policy_manager.go
//tests the extensions policy manager configuration
//test for purpose "fraud-detection", connector mocks configured for specific output per purpose
func TestExtPolicyManagerPurpose1(t *testing.T) {
	policyDecision := policyManagerExtOnly("fraud-detection")
	extPMpolicies := tu.GetExtPMDecisions("fraud-detection")
	tu.EnsureDeepEqualDecisions(t, policyDecision, extPMpolicies)
}

//Tests GetEnforcementActions func in policy_manager.go
//tests the extensions policy manager configuration
//test for purpose "marketing", connector mocks configured for specific output per purpose
func TestExtPolicyManagerPurpose2(t *testing.T) {
	policyDecision := policyManagerExtOnly("marketing")
	extPMpolicies := tu.GetExtPMDecisions("marketing")
	tu.EnsureDeepEqualDecisions(t, policyDecision, extPMpolicies)
}
