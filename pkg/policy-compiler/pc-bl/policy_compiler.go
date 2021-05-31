// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policycompilerbl

import (
	"encoding/json"
	"fmt"
	"log"

	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
)

// PolicyCompiler iassumed to have main policies-manager connector and may have or not extension policies-manager
type PolicyCompiler struct {
	mainPolicyManagerCon       PolicyManagerConnector
	extensionsPolicyManagerCon PolicyManagerConnector
	useExtensionPC             bool
}

func NewPolicyCompiler(mainPolicyManagerCon, extensionsPolicyManagerCon PolicyManagerConnector, useExtensionPC bool) PolicyCompiler {
	return PolicyCompiler{mainPolicyManagerCon: mainPolicyManagerCon,
		extensionsPolicyManagerCon: extensionsPolicyManagerCon, useExtensionPC: useExtensionPC}
}

func NewPolicyCompilerSingleCon(mainPolicyManagerCon PolicyManagerConnector) PolicyCompiler {
	return PolicyCompiler{mainPolicyManagerCon: mainPolicyManagerCon, useExtensionPC: false}
}

func (pc PolicyCompiler) GetEnforcementActions(appContext *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
	fmt.Println("*********************************in Policy Compiler GetEnforcementActions *****************************")

	var policiesDecisions *pb.PoliciesDecisions

	policiesDecisions, err := pc.mainPolicyManagerCon.GetEnforcementActions(appContext)
	if err != nil {
		log.Printf("Error Fetching Response from %s Connector: %v", pc.mainPolicyManagerCon.name, err)

		// propagating the error with the relevant grpc error codes and description back to m4d pilot
		return nil, err
	}

	jsonOutput, _ := json.MarshalIndent(policiesDecisions, "", "\t")
	log.Printf("Enforcement Actions from %s: %v\n", pc.mainPolicyManagerCon.name, string(jsonOutput))

	if pc.useExtensionPC {
		policiesDecisionsExt, err := pc.extensionsPolicyManagerCon.GetEnforcementActions(appContext)

		if err != nil {
			log.Printf("Error Fetching Response from %s Connector: %v\n", pc.extensionsPolicyManagerCon.name, err)

			// propagating the error with the relevant grpc error codes and description back to m4d pilot
			return nil, err
		}
		jsonOutput, _ = json.MarshalIndent(policiesDecisionsExt, "", "\t")
		log.Printf("Enforcement Actions from %s: %v\n", pc.extensionsPolicyManagerCon.name, string(jsonOutput))

		policiesDecisions = GetCombinedPoliciesDecisions(policiesDecisions, policiesDecisionsExt)
	}

	serverComponentVersion := &pb.ComponentVersion{Name: "PCComponent", Id: "ID-1", Version: "1.0"}
	policiesDecisions.ComponentVersions = append(policiesDecisions.ComponentVersions, serverComponentVersion)

	log.Println("********************************* Combined Policies *****************************")

	jsonOutput, _ = json.MarshalIndent(policiesDecisions, "", "\t")
	log.Println(string(jsonOutput))

	log.Println("*********************************************************************************")

	return policiesDecisions, nil
}
