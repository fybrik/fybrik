// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policycompiler

import (
	"fmt"
	"os"
	"testing"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	bl "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/pc-bl"
	tu "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/testutil"
)

func constructPolicyConnectors() (*bl.PolicyManagerConnector, *bl.PolicyManagerConnector) {
	mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs, extensionPolicyManagerName, extensionPolicyManagerURL := tu.GetEnvironment()

	mainPolicyManager := bl.NewPolicyManagerConnector(mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs)
	extPolicyManager := bl.NewPolicyManagerConnector(extensionPolicyManagerName, extensionPolicyManagerURL, timeOutSecs)

	return &mainPolicyManager, &extPolicyManager
}

func verifyWrapperPolicyCompiler(t *testing.T, policyCompiler *bl.PolicyCompiler, applicationContext *pb.ApplicationContext) {
	innerDecisions, _ := policyCompiler.GetEnforcementActions(applicationContext)
	pcSvc := &PolicyCompiler{policyCompiler: *policyCompiler}
	decisions, _ := pcSvc.GetPoliciesDecisions(applicationContext)

	tu.EnsureDeepEqualDecisions(t, decisions, innerDecisions)
}

//Tests  GetPoliciesDecisions in policy_compiler_server.go
//tests the extension policy manager configuration (used instead on main policy-maanager)
//test for purpose "fraud-detection" and "marketing" purposes, connector mocks configured for different outputs for these purposes
func TestExtPolicyCompiler(t *testing.T) {
	mainPolicyManager, extPolicyManager := constructPolicyConnectors()

	applicationContext := tu.GetApplicationContext("fraud-detection")
	policyCompiler := bl.NewPolicyCompiler(*mainPolicyManager, *extPolicyManager, true)
	verifyWrapperPolicyCompiler(t, &policyCompiler, applicationContext)
}

func TestExtPolicyCompilerReversed(t *testing.T) {
	mainPolicyManager, extPolicyManager := constructPolicyConnectors()

	applicationContext := tu.GetApplicationContext("marketing")
	policyCompiler := bl.NewPolicyCompiler(*extPolicyManager, *mainPolicyManager, true)
	verifyWrapperPolicyCompiler(t, &policyCompiler, applicationContext)
}

func TestMain(m *testing.M) {
	fmt.Println("TestMain function called = policy_compiler_server_test ")

	tu.EnvValues["MAIN_POLICY_MANAGER_URL"] = "localhost:" + "40082"
	tu.EnvValues["EXTENSIONS_POLICY_MANAGER_URL"] = "localhost:" + "40092"

	go tu.MockMainConnector(40082)
	go tu.MockExtConnector(40092)
	code := m.Run()
	fmt.Println("TestMain function called after Run = policy_compiler_server_test ")
	os.Exit(code)
}
