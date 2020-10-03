// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policycompilerbl

import (
	"fmt"
	"os"
	"testing"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	tu "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/testutil"
)

func policyCompilerMainOnly(purpose string) *pb.PoliciesDecisions {
	mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs, _, _ := tu.GetEnvironment()
	applicationContext := tu.GetApplicationContext(purpose)

	mainPolicyManager := NewPolicyManagerConnector(mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs)

	policyCompiler := NewPolicyCompilerSingleCon(mainPolicyManager)
	policies, _ := policyCompiler.GetEnforcementActions(applicationContext)
	return policies
}

func policyCompilerExtOnly(purpose string) *pb.PoliciesDecisions {
	_, _, timeOutSecs, extensionPolicyManagerName, extensionPolicyManagerURL := tu.GetEnvironment()
	applicationContext := tu.GetApplicationContext(purpose)

	extPolicyManager := NewPolicyManagerConnector(extensionPolicyManagerName, extensionPolicyManagerURL, timeOutSecs)
	policyCompiler := NewPolicyCompilerSingleCon(extPolicyManager)
	policies, _ := policyCompiler.GetEnforcementActions(applicationContext)
	return policies
}

func policyCompilerMainAndExt(purpose string) *pb.PoliciesDecisions {
	mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs, extensionPolicyManagerName, extensionPolicyManagerURL := tu.GetEnvironment()
	applicationContext := tu.GetApplicationContext(purpose)

	mainPolicyManager := NewPolicyManagerConnector(mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs)
	extPolicyManager := NewPolicyManagerConnector(extensionPolicyManagerName, extensionPolicyManagerURL, timeOutSecs)

	policyCompiler := NewPolicyCompiler(mainPolicyManager, extPolicyManager, true)
	policies, _ := policyCompiler.GetEnforcementActions(applicationContext)
	return policies
}

func policyCompilerMainAndExtReversed(purpose string) *pb.PoliciesDecisions {
	mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs, extensionPolicyManagerName, extensionPolicyManagerURL := tu.GetEnvironment()
	applicationContext := tu.GetApplicationContext(purpose)

	mainPolicyManager := NewPolicyManagerConnector(mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs)
	extPolicyManager := NewPolicyManagerConnector(extensionPolicyManagerName, extensionPolicyManagerURL, timeOutSecs)

	policyCompiler := NewPolicyCompiler(extPolicyManager, mainPolicyManager, true)
	policies, _ := policyCompiler.GetEnforcementActions(applicationContext)
	return policies
}

//Tests  GetEnforcementActions in policy_compiler.go
//tests the main policy manager configuration
//test for purpose "fraud-detection" and "marketing" purposes, connector mocks configured for different outputs for these purposes
func TestMainPolicyCompiler(t *testing.T) {
	policyDecision := policyCompilerMainOnly("fraud-detection")
	mainPMpolicies := tu.GetMainPMDecisions("fraud-detection")
	//we add component version similar to the one added by PolicyCompiler
	serverComponentVersion := &pb.ComponentVersion{Name: "PCComponent", Id: "ID-1", Version: "1.0"}
	mainPMpolicies.ComponentVersions = append(mainPMpolicies.ComponentVersions, serverComponentVersion)
	tu.EnsureDeepEqualDecisions(t, policyDecision, mainPMpolicies)

	policyDecision2 := policyCompilerMainOnly("marketing")
	mainPMpolicies2 := tu.GetMainPMDecisions("marketing")
	mainPMpolicies2.ComponentVersions = append(mainPMpolicies2.ComponentVersions, serverComponentVersion)
	tu.EnsureDeepEqualDecisions(t, policyDecision2, mainPMpolicies2)
}

//Tests  GetEnforcementActions in policy_compiler.go
//tests the extension policy manager configuration (used instead on main policy-maanager)
//test for purpose "fraud-detection" and "marketing" purposes, connector mocks configured for different outputs for these purposes
func TestExtPolicyCompiler(t *testing.T) {
	policyDecision := policyCompilerExtOnly("fraud-detection")
	extPMpolicies := tu.GetExtPMDecisions("fraud-detection")
	//we add component version similar to the one added by PolicyCompiler
	serverComponentVersion := &pb.ComponentVersion{Name: "PCComponent", Id: "ID-1", Version: "1.0"}
	extPMpolicies.ComponentVersions = append(extPMpolicies.ComponentVersions, serverComponentVersion)
	tu.EnsureDeepEqualDecisions(t, policyDecision, extPMpolicies)

	policyDecision2 := policyCompilerExtOnly("marketing")
	extPMpolicies2 := tu.GetExtPMDecisions("marketing")
	extPMpolicies2.ComponentVersions = append(extPMpolicies2.ComponentVersions, serverComponentVersion)
	tu.EnsureDeepEqualDecisions(t, policyDecision2, extPMpolicies2)
}

//Tests  GetEnforcementActions in policy_compiler.go
//tests main policy manager configuration combined with extension policy manager
//test for purpose "fraud-detection" and "marketing" purposes, connector mocks configured for different outputs for these purposes
func TestMainAndExtPolicyCompiler(t *testing.T) {
	policyDecision := policyCompilerMainAndExt("fraud-detection")
	mainPMpolicies := tu.GetMainPMDecisions("fraud-detection")
	extPMpolicies := tu.GetExtPMDecisions("fraud-detection")
	tu.CheckPolicies(t, policyDecision, mainPMpolicies, extPMpolicies)

	policyDecision2 := policyCompilerMainAndExt("marketing")
	mainPMpolicies2 := tu.GetMainPMDecisions("marketing")
	extPMpolicies2 := tu.GetExtPMDecisions("marketing")
	tu.CheckPolicies(t, policyDecision2, mainPMpolicies2, extPMpolicies2)
}

//Tests  GetEnforcementActions in policy_compiler.go
//tests main policy manager configuration combined with extension policy manager - after we switch the main and the extension one
//test for purpose "fraud-detection" and "marketing" purposes, connector mocks configured for different outputs for these purposes
func TestMainAndExtPolicyCompilerReversed(t *testing.T) {
	policyDecision := policyCompilerMainAndExtReversed("fraud-detection")
	mainPMpolicies := tu.GetMainPMDecisions("fraud-detection")
	extPMpolicies := tu.GetExtPMDecisions("fraud-detection")
	tu.CheckPolicies(t, policyDecision, mainPMpolicies, extPMpolicies)

	policyDecision2 := policyCompilerMainAndExtReversed("marketing")
	mainPMpolicies2 := tu.GetMainPMDecisions("marketing")
	extPMpolicies2 := tu.GetExtPMDecisions("marketing")
	tu.CheckPolicies(t, policyDecision2, mainPMpolicies2, extPMpolicies2)
}

//TestMain executes the above defined test function.

func TestMain(m *testing.M) {
	fmt.Println("TestMain function called = policy_compiler_test ")

	tu.EnvValues["MAIN_POLICY_MANAGER_URL"] = "localhost:" + "40081"
	tu.EnvValues["EXTENSIONS_POLICY_MANAGER_URL"] = "localhost:" + "40091"

	go tu.MockMainConnector(40081)
	go tu.MockExtConnector(40091)
	code := m.Run()
	fmt.Println("TestMain function called after Run = policy_compiler_test ")
	os.Exit(code)
}
