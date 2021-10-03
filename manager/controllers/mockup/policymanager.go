// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	connectors "fybrik.io/fybrik/pkg/connectors/clients"
	"fybrik.io/fybrik/pkg/random"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
)

// MockPolicyManager is a mock for PolicyManager interface used in tests
type MockPolicyManager struct {
	connectors.PolicyManager
}

// GetPoliciesDecisions implements the PolicyCompiler interface
func (m *MockPolicyManager) GetPoliciesDecisions(input *openapiclientmodels.PolicyManagerRequest, creds string) (*openapiclientmodels.PolicyManagerResponse, error) {
	log.Printf("Received OpenAPI request in mockup GetPoliciesDecisions: ")
	log.Printf("ProcessingGeography: " + input.Action.GetProcessingLocation())
	log.Printf("Destination: " + *input.Action.Destination)

	datasetID := input.GetResource().Name
	log.Printf("   DataSetID: " + datasetID)
	respResult := []openapiclientmodels.ResultItem{}
	policyManagerResult := openapiclientmodels.ResultItem{}

	splittedID := strings.SplitN(datasetID, "/", 2)
	if len(splittedID) != 2 {
		panic(fmt.Sprintf("Invalid dataset ID for mock: %s", datasetID))
	}
	assetID := splittedID[1]
	switch assetID {
	case "allow-dataset":
		actionOnDataset := openapiclientmodels.Action{}
		(&actionOnDataset).SetName("Allow")
		policyManagerResult.SetAction(actionOnDataset)
	case "deny-dataset":
		actionOnDataset := openapiclientmodels.Action{}
		(&actionOnDataset).SetName("Deny")
		policyManagerResult.SetAction(actionOnDataset)
	case "allow-theshire":
		actionOnDataset := openapiclientmodels.Action{}
		if *input.GetAction().Destination == "theshire" {
			(&actionOnDataset).SetName("Allow")
		} else {
			(&actionOnDataset).SetName("Deny")
		}
		policyManagerResult.SetAction(actionOnDataset)
	case "deny-theshire":
		actionOnDataset := openapiclientmodels.Action{}
		if *input.GetAction().Destination != "theshire" {
			(&actionOnDataset).SetName("Allow")
		} else {
			(&actionOnDataset).SetName("Deny")
		}
		policyManagerResult.SetAction(actionOnDataset)
	default:
		actionOnCols := openapiclientmodels.Action{}
		action := make(map[string]interface{})
		action["name"] = "redact"
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
	}
	respResult = append(respResult, policyManagerResult)
	decisionID, _ := random.Hex(20)
	policyManagerResp := &openapiclientmodels.PolicyManagerResponse{DecisionId: &decisionID, Result: respResult}

	res, err := json.MarshalIndent(policyManagerResp, "", "\t")
	if err != nil {
		log.Println("error in marshalling policy manager response :", err)
		return nil, err
	}
	log.Println("Marshalled policy manager response in mockup GetPoliciesDecisions:", string(res))

	return policyManagerResp, nil
}
