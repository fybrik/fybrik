// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	pmclient "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	"fybrik.io/fybrik/pkg/random"
	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
)

// MockPolicyManager is a mock for PolicyManager interface used in tests
type MockPolicyManager struct {
	pmclient.PolicyManager
}

// GetPoliciesDecisions implements the PolicyCompiler interface
func (m *MockPolicyManager) GetPoliciesDecisions(input *taxonomymodels.PolicyManagerRequest, creds string) (*taxonomymodels.PolicyManagerResponse, error) {
	log.Printf("Received OpenAPI request in mockup GetPoliciesDecisions: ")
	log.Printf("ProcessingGeography: " + input.Action.GetProcessingLocation())
	log.Printf("Destination: " + *input.Action.Destination)

	datasetID := input.GetResource().Name
	log.Printf("   DataSetID: " + datasetID)
	respResult := []taxonomymodels.PolicyManagerResultItem{}
	policyManagerResult := taxonomymodels.PolicyManagerResultItem{}

	splittedID := strings.SplitN(datasetID, "/", 2)
	if len(splittedID) != 2 {
		panic(fmt.Sprintf("Invalid dataset ID for mock: %s", datasetID))
	}
	assetID := splittedID[1]
	switch assetID {
	case "allow-dataset":
		// empty result simulates allow
		// no need to construct any result item
	case "deny-dataset":
		actionOnDataset := taxonomymodels.Action{}
		(&actionOnDataset).SetName("Deny")
		policyManagerResult.SetAction(actionOnDataset)
		respResult = append(respResult, policyManagerResult)
	case "allow-theshire":
		if *input.GetAction().Destination != "theshire" {
			actionOnDataset := taxonomymodels.Action{}
			(&actionOnDataset).SetName("Deny")
			policyManagerResult.SetAction(actionOnDataset)
			respResult = append(respResult, policyManagerResult)
		}
	case "deny-theshire":
		if *input.GetAction().Destination == "theshire" {
			actionOnDataset := taxonomymodels.Action{}
			(&actionOnDataset).SetName("Deny")
			policyManagerResult.SetAction(actionOnDataset)
			respResult = append(respResult, policyManagerResult)
		}
	default:
		actionOnCols := taxonomymodels.Action{}
		action := make(map[string]interface{})
		action["name"] = "RedactAction"
		action["column"] = []string{"SSN"}

		actionBytes, errJSON := json.MarshalIndent(action, "", "\t")
		if errJSON != nil {
			return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
		}
		err := json.Unmarshal(actionBytes, &actionOnCols)
		if err != nil {
			return nil, fmt.Errorf("error in unmarshalling actionBytes : %v", err)
		}
		policyManagerResult.SetAction(actionOnCols)
		respResult = append(respResult, policyManagerResult)
	}

	decisionID, _ := random.Hex(20)
	policyManagerResp := &taxonomymodels.PolicyManagerResponse{DecisionId: &decisionID, Result: respResult}

	res, err := json.MarshalIndent(policyManagerResp, "", "\t")
	if err != nil {
		log.Println("error in marshalling policy manager response :", err)
		return nil, err
	}
	log.Println("Marshalled policy manager response in mockup GetPoliciesDecisions:", string(res))

	return policyManagerResp, nil
}
