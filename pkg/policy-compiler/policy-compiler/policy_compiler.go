// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policycompiler

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	bl "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/pc-bl"
)

type PolicyCompiler struct {
	IPolicyCompiler
	policyCompiler bl.PolicyCompiler
}

type IPolicyCompiler interface {
	GetPoliciesDecisions(in *pb.ApplicationContext) (*pb.PoliciesDecisions, error)
}

func NewPolicyCompiler() *PolicyCompiler {
	pc := &PolicyCompiler{}
	pc.initialize()
	return pc
}

func (s *PolicyCompiler) GetPoliciesDecisions(in *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
	log.Println("*********************************Received ApplicationContext *****************************")
	log.Println(in)
	log.Println("******************************************************************************************")

	eval, err := s.policyCompiler.GetEnforcementActions(in)
	jsonOutput, _ := json.MarshalIndent(eval, "", "\t")
	log.Println("Received evaluation : " + string(jsonOutput))
	log.Println("err:", err)

	return eval, err
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Env Variable %v not defined", key)
	}
	return value
}

func (s *PolicyCompiler) initialize() {
	timeOutInSecs := getEnv("CONNECTION_TIMEOUT")
	timeOutSecs, err := strconv.Atoi(timeOutInSecs)
	if err != nil {
		log.Fatalf("Atoi conversion of timeOutinseconds failed: %v", err)
	} else {
		log.Println("timeOut env variable in PolicyCompilerServer: ", timeOutSecs)
	}

	mainPolicyManagerURL := getEnv("MAIN_POLICY_MANAGER_CONNECTOR_URL")
	mainPolicyManagerName := getEnv("MAIN_POLICY_MANAGER_NAME")
	mainPolicyManager := bl.NewPolicyManagerConnector(mainPolicyManagerName, mainPolicyManagerURL, timeOutSecs)
	log.Printf("PolicyCompilerServer: main policy manager name %s, URL: %s\n", mainPolicyManagerName, mainPolicyManagerURL)

	var extensionPolicyManager bl.PolicyManagerConnector
	useExtensionPolicies := getEnv("USE_EXTENSIONPOLICY_MANAGER")
	useExtensionsPC, _ := strconv.ParseBool(useExtensionPolicies)
	if useExtensionsPC {
		extensionPolicyManagerURL := getEnv("EXTENSIONS_POLICY_MANAGER_CONNECTOR_URL")
		extensionPolicyManagerName := getEnv("EXTENSIONS_POLICY_MANAGER_NAME")
		extensionPolicyManager = bl.NewPolicyManagerConnector(extensionPolicyManagerName, extensionPolicyManagerURL, timeOutSecs)
		log.Printf("PolicyCompilerServer: extensions policy manager is used, name: %s, URL: %s\n", extensionPolicyManagerName, extensionPolicyManagerURL)
	} else {
		log.Println("PolicyCompilerServer: no extension policy manager is used")
	}

	s.policyCompiler = bl.NewPolicyCompiler(mainPolicyManager, extensionPolicyManager, useExtensionsPC)
}
