// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
)

type OpaReader struct {
	opaServerURL string
}

func NewOpaReader(opasrvurl string) *OpaReader {
	return &OpaReader{opaServerURL: opasrvurl}
}

func (r *OpaReader) GetOPADecisions(in *pb.ApplicationContext, catalogReader *CatalogReader, policyToBeEvaluated string) (*pb.PoliciesDecisions, error) {
	datasetsMetadata, err := catalogReader.GetDatasetsMetadataFromCatalog(in)
	if err != nil {
		return nil, err
	}

	appInfo := in.GetAppInfo()
	appInfoBytes, err := json.MarshalIndent(appInfo, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("error in marshalling appInfo: %v", err)
	}
	log.Println("appInfo : " + string(appInfoBytes))
	appInfoMap := make(map[string]interface{})
	err = json.Unmarshal(appInfoBytes, &appInfoMap)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshalling appInfoBytes: %v", err)
	}

	// to store the list of DatasetDecision
	var datasetDecisionList []*pb.DatasetDecision
	for i, datasetContext := range in.GetDatasets() {
		dataset := datasetContext.GetDataset()
		datasetID := dataset.GetDatasetId()
		metadata := datasetsMetadata[datasetID]

		inputMap, ok := metadata.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error in unmarshalling dataset metadata (datasetID = %s): %v", datasetID, err)
		}

		operation := datasetContext.GetOperation()
		// Encode operation in a map[string]interface
		operationBytes, err := json.MarshalIndent(operation, "", "\t")
		log.Println("Operation Bytes: " + string(operationBytes))
		if err != nil {
			return nil, fmt.Errorf("error in marshalling operation (i = %d): %v", i, err)
		}
		operationMap := make(map[string]interface{})
		err = json.Unmarshal(operationBytes, &operationMap)
		if err != nil {
			return nil, fmt.Errorf("error in marshalling into operation map (i = %d): %v", i, err)
		}
		for k, v := range operationMap {
			if k == "type" {
				inputMap[k] = pb.AccessOperation_AccessType_name[int32(operation.GetType())]
			} else {
				inputMap[k] = v
			}
		}
		// Combine with appInfoMap
		for k, v := range appInfoMap {
			inputMap[k] = v
		}
		// Printing the combined map
		toPrintBytes, _ := json.MarshalIndent(inputMap, "", "\t")
		log.Println("********sending this to OPA : *******")
		log.Println(string(toPrintBytes))
		opaEval, err := EvaluatePoliciesOnInput(inputMap, r.opaServerURL, policyToBeEvaluated)
		if err != nil {
			log.Printf("error in EvaluatePoliciesOnInput (i = %d): %v", i, err)
			return nil, fmt.Errorf("error in EvaluatePoliciesOnInput (i = %d): %v", i, err)
		}
		log.Println("OPA Eval : " + opaEval)
		opaOperationDecision, err := GetOPAOperationDecision(opaEval, operation)
		if err != nil {
			return nil, fmt.Errorf("error in GetOPAOperationDecision (i = %d): %v", i, err)
		}
		// Add to a list
		var opaOperationDecisionList []*pb.OperationDecision
		opaOperationDecisionList = append(opaOperationDecisionList, opaOperationDecision)
		// Create a new *DatasetDecision
		datasetDecison := &pb.DatasetDecision{Dataset: dataset, Decisions: opaOperationDecisionList}
		datasetDecisionList = append(datasetDecisionList, datasetDecison)
	}
	return &pb.PoliciesDecisions{DatasetDecisions: datasetDecisionList}, nil
}

// Translate the evaluation received from OPA for (dataset, operation) into pb.OperationDecision
func GetOPAOperationDecision(opaEval string, operation *pb.AccessOperation) (*pb.OperationDecision, error) {
	resultInterface := make(map[string]interface{})
	err := json.Unmarshal([]byte(opaEval), &resultInterface)
	if err != nil {
		return nil, err
	}
	evaluationMap, ok := resultInterface["result"].(map[string]interface{})
	if !ok {
		return nil, errors.New("error in format of OPA evaluation (incorrect result map)")
	}

	// Now iterate over
	enforcementActions := make([]*pb.EnforcementAction, 0)
	usedPolicies := make([]*pb.Policy, 0)

	if evaluationMap["deny"] != nil {
		lstDeny, ok := evaluationMap["deny"].([]interface{})
		if !ok {
			return nil, errors.New("unknown format of deny content")
		}
		if len(lstDeny) > 0 {
			newEnforcementAction := &pb.EnforcementAction{Name: "Deny", Id: "Deny-ID", Level: pb.EnforcementAction_DATASET, Args: map[string]string{}}
			enforcementActions = append(enforcementActions, newEnforcementAction)

			for i, reason := range lstDeny {
				if reasonMap, ok := reason.(map[string]interface{}); ok {
					if newUsedPolicy, ok := buildNewPolicy(reasonMap["used_policy"]); ok {
						usedPolicies = append(usedPolicies, newUsedPolicy)
						continue
					}
				}
				log.Printf("Warning: unknown format of argument %d of lstDeny list. Skipping", i)
				continue
			}
		}
	}

	if evaluationMap["transform"] != nil {
		lstTransformations, ok := evaluationMap["transform"].([]interface{})
		if !ok {
			return nil, errors.New("unknown format of transform content")
		}
		for i, transformAction := range lstTransformations {
			newEnforcementAction, newUsedPolicy, ok := buildNewEnfrocementAction(transformAction)
			if !ok {
				return nil, errors.New("unknown format of transform action")
			}
			enforcementActions = append(enforcementActions, newEnforcementAction)
			if newUsedPolicy == nil {
				log.Printf("Warning: empty used policy field for transformation %d", i)
			} else {
				usedPolicies = append(usedPolicies, newUsedPolicy)
			}
		}
	}

	if len(enforcementActions) == 0 { // allow action
		newEnforcementAction := &pb.EnforcementAction{Name: "Allow", Id: "Allow-ID", Level: pb.EnforcementAction_DATASET, Args: map[string]string{}}
		enforcementActions = append(enforcementActions, newEnforcementAction)
	}

	log.Println("enforcementActions: ", enforcementActions)
	log.Println("usedPolicies: ", usedPolicies)

	return &pb.OperationDecision{Operation: operation, EnforcementActions: enforcementActions, UsedPolicies: usedPolicies}, nil
}

func buildNewEnfrocementAction(transformAction interface{}) (*pb.EnforcementAction, *pb.Policy, bool) {
	if action, ok := transformAction.(map[string]interface{}); ok {
		newUsedPolicy, ok := buildNewPolicy(action["used_policy"])
		if !ok {
			log.Println("Warning: unknown format of used policy information. Skipping policy", action)
		}

		if result, ok := action["action_name"].(string); ok {
			switch result {
			case "remove column":
				if columnName, ok := extractArgument(action["arguments"], "column_name"); ok {
					newEnforcementAction := &pb.EnforcementAction{Name: "removed", Id: "removed-ID",
						Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": columnName}}
					return newEnforcementAction, newUsedPolicy, true
				}
			case "encrypt column":
				if columnName, ok := extractArgument(action["arguments"], "column_name"); ok {
					newEnforcementAction := &pb.EnforcementAction{Name: "encrypted", Id: "encrypted-ID",
						Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": columnName}}
					return newEnforcementAction, newUsedPolicy, true
				}
			case "redact column":
				if columnName, ok := extractArgument(action["arguments"], "column_name"); ok {
					newEnforcementAction := &pb.EnforcementAction{Name: "redact", Id: "redact-ID",
						Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": columnName}}
					return newEnforcementAction, newUsedPolicy, true
				}
			case "periodic blackout":
				if monthlyDaysNum, ok := extractArgument(action["arguments"], "monthly_days_end"); ok {
					newEnforcementAction := &pb.EnforcementAction{Name: "periodic_blackout", Id: "periodic_blackout-ID",
						Level: pb.EnforcementAction_DATASET, Args: map[string]string{"monthly_days_end": monthlyDaysNum}}
					return newEnforcementAction, newUsedPolicy, true
				} else if yearlyDaysNum, ok := extractArgument(action["arguments"], "yearly_days_end"); ok {
					newEnforcementAction := &pb.EnforcementAction{Name: "periodic_blackout", Id: "periodic_blackout-ID",
						Level: pb.EnforcementAction_DATASET, Args: map[string]string{"yearly_days_end": yearlyDaysNum}}
					return newEnforcementAction, newUsedPolicy, true
				}
			default:
				log.Printf("Unknown Enforcement Action receieved from OPA")
			}
		}
	}
	return nil, nil, false
}

func extractArgument(arguments interface{}, argName string) (string, bool) {
	if argsMap, ok := arguments.(map[string]interface{}); ok {
		if value, ok := argsMap[argName].(string); ok {
			return value, true
		}
	}
	return "", false
}

func buildNewPolicy(usedPolicy interface{}) (*pb.Policy, bool) {
	if policy, ok := usedPolicy.(map[string]interface{}); ok {
		//todo: add other fields that can be returned as part of the policy struct
		if description, ok := policy["description"].(string); ok {
			newUsedPolicy := &pb.Policy{Description: description}
			return newUsedPolicy, true
		}
	}

	return nil, false
}
