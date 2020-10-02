// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package opaconnbl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	opaServerURL string
}

func NewServer(opasrvurl string) *Server {
	return &Server{opaServerURL: opasrvurl}
}

func (s *Server) GetPoliciesDecisions(in *pb.ApplicationContext, catalogConnectorAddress string, timeOut int) (*pb.PoliciesDecisions, error) {

	log.Println("Using catalog connector address: ", catalogConnectorAddress)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, catalogConnectorAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("Connection to External Catalog Connector failed: %v", err)
		errStatus, _ := status.FromError(err)
		log.Println(errStatus.Message())
		log.Println(errStatus.Code())
		return nil, fmt.Errorf("Connection to External Catalog Connector failed: %v", err)
	}
	defer conn.Close()
	c := pb.NewDataCatalogServiceClient(conn)
	//Encode appInfo in a map[string]interface
	appID := in.GetAppId()
	appInfo := in.GetAppInfo()
	appInfoBytes, err := json.MarshalIndent(appInfo, "", "\t")
	if err != nil {
		log.Printf("error in marshalling appInfo: %v", err)
		return nil, fmt.Errorf("error in marshalling appInfo: %v", err)
	}
	log.Println("appInfo : " + string(appInfoBytes))
	appInfoMap := make(map[string]interface{})
	err = json.Unmarshal(appInfoBytes, &appInfoMap)
	if err != nil {
		log.Printf("error in unmarshalling appInfoBytes: %v", err)
		return nil, fmt.Errorf("error in unmarshalling appInfoBytes: %v", err)
	}
	//to store the list of DatasetDecision
	var datasetDecisionList []*pb.DatasetDecision
	for i, datasetContext := range in.GetDatasets() {
		dataset := datasetContext.GetDataset()
		datasetID := dataset.GetDatasetId()
		inputMap, err := GetDatasetMetadata(datasetID, appID, &ctx, c, i)
		if err != nil {
			return nil, err
		}
		operation := datasetContext.GetOperation()
		//Encode operation in a map[string]interface
		operationBytes, err := json.MarshalIndent(operation, "", "\t")
		log.Println("Operation Bytes: " + string(operationBytes))
		if err != nil {
			log.Printf("error in marshalling operation (i = %d): %v", i, err)
			return nil, fmt.Errorf("error in marshalling operation (i = %d): %v", i, err)
		}
		operationMap := make(map[string]interface{})
		err = json.Unmarshal(operationBytes, &operationMap)
		if err != nil {
			log.Printf("error in marshalling into operation map (i = %d): %v", i, err)
			return nil, fmt.Errorf("error in marshalling into operation map (i = %d): %v", i, err)
		}
		for k, v := range operationMap {
			if k == "type" {
				inputMap[k] = pb.AccessOperation_AccessType_name[int32(operation.GetType())]
			} else {
				inputMap[k] = v
			}
		}
		//Combine with appInfoMap
		for k, v := range appInfoMap {
			inputMap[k] = v
		}
		//Printing the combined map
		toPrintBytes, _ := json.MarshalIndent(inputMap, "", "\t")
		log.Println("********sending this to OPA : *******")
		log.Println(string(toPrintBytes))
		opaEval, err := EvaluateExtendedPoliciesOnInput(inputMap, s.opaServerURL)
		if err != nil {
			log.Printf("error in EvaluateExtendedPoliciesOnInput (i = %d): %v", i, err)
			return nil, fmt.Errorf("error in EvaluateExtendedPoliciesOnInput (i = %d): %v", i, err)
		}
		log.Println("OPA Eval : " + opaEval)
		opaOperationDecision, err := GetOPAOperationDecision(opaEval, operation)
		if err != nil {
			log.Printf("error in GetOPAOperationDecision (i = %d): %v", i, err)
			return nil, fmt.Errorf("error in GetOPAOperationDecision (i = %d): %v", i, err)
		}
		//add to a list
		var opaOperationDecisionList []*pb.OperationDecision
		opaOperationDecisionList = append(opaOperationDecisionList, opaOperationDecision)
		//Create a new *DatasetDecision
		datasetDecison := &pb.DatasetDecision{Dataset: dataset, Decisions: opaOperationDecisionList}
		datasetDecisionList = append(datasetDecisionList, datasetDecison)
	}
	return &pb.PoliciesDecisions{DatasetDecisions: datasetDecisionList}, nil
}

//Translate the evaluation received from OPA for (dataset, operation) into pb.OperationDecision
func GetOPAOperationDecision(opaEval string, operation *pb.AccessOperation) (*pb.OperationDecision, error) {
	resultInterface := make(map[string]interface{})
	err := json.Unmarshal([]byte(opaEval), &resultInterface)
	if err != nil {
		log.Printf("error in unmarshaling opaEval into resultInterface: " + err.Error())
		return nil, err
	}
	mainMap, ok := resultInterface["result"].(map[string]interface{})
	if !ok {
		log.Printf("error in format of OPA evaluation (incorrect result map)")
		return nil, errors.New("error in format of OPA evaluation (incorrect result map)")
	}

	// log.Printf("************************** input: ", mainMap)

	//Now iterate over
	enforcementActions := make([]*pb.EnforcementAction, 0)
	usedPolicies := make([]*pb.Policy, 0)

	if mainMap["deny"] != nil {
		lstDeny, ok := mainMap["deny"].([]interface{})
		if !ok {
			log.Printf("Error: unknown format of deny list")
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

	if mainMap["transform"] != nil {
		lstTransformations, ok := mainMap["transform"].([]interface{})
		if !ok {
			log.Printf("Error: unknown format of transformationss list")
			return nil, errors.New("unknown format of transform content")
		}
		for i, transformAction := range lstTransformations {
			newEnforcementAction, newUsedPolicy, ok := buildNewEnfrocementAction(transformAction)
			if !ok {
				log.Printf("Error: unknown format of transform action")
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

	if len(enforcementActions) == 0 { //allow action
		newEnforcementAction := &pb.EnforcementAction{Name: "Allow", Id: "Allow-ID", Level: pb.EnforcementAction_DATASET, Args: map[string]string{}}
		enforcementActions = append(enforcementActions, newEnforcementAction)
	}

	// for k, v := range mainMap {
	// 	switch k {
	// 	case "Encrypt_col":
	// 		newEnforcementActions, newUsedPolicies := GetEncryptColEnforcementActionsAndPolicies(v)
	// 		enforcementActions = append(enforcementActions, newEnforcementActions...)
	// 		usedPolicies = append(usedPolicies, newUsedPolicies...)
	// 	default:
	// 		log.Printf("Unknown Enforcement Action receieved from OPA")
	// 	}
	// }

	log.Println("************************** EA: ", enforcementActions)
	log.Println("************************** POL: ", usedPolicies)

	return &pb.OperationDecision{Operation: operation, EnforcementActions: enforcementActions, UsedPolicies: usedPolicies}, nil
}

func buildNewEnfrocementAction(transformAction interface{}) (*pb.EnforcementAction, *pb.Policy, bool) {
	if action, ok := transformAction.(map[string]interface{}); ok {
		newUsedPolicy, ok := buildNewPolicy(action["used_policy"])
		if !ok {
			log.Println("Warning: unknown format of used policy information. Skipping policy", action)
		}

		if result, ok := action["result"].(string); ok {
			switch result {
			case "Remove column":
				if columnName, ok := extractArgument(action["args"], "column name"); ok {
					newEnforcementAction := &pb.EnforcementAction{Name: "removed", Id: "removed-ID",
						Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": columnName}}
					return newEnforcementAction, newUsedPolicy, true
				}
			case "Redact column":
				if columnName, ok := extractArgument(action["args"], "column name"); ok {
					newEnforcementAction := &pb.EnforcementAction{Name: "redact", Id: "redact-ID",
						Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": columnName}}
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

func GetDatasetMetadata(datasetID string, appID string, ctx *context.Context, c pb.DataCatalogServiceClient, i int) (map[string]interface{}, error) {
	objToSend := &pb.CatalogDatasetRequest{AppId: appID, DatasetId: datasetID}
	log.Println("Sending request to External Catalog Connector")
	r, err := c.GetDatasetInfo(*ctx, objToSend)
	if err != nil {
		log.Printf("error sending data to External Catalog Connector (i = %d): %v", i, err)
		errStatus, _ := status.FromError(err)
		log.Println("Message:", errStatus.Message())
		log.Println("Code:", errStatus.Code())
		if codes.InvalidArgument == errStatus.Code() {
			log.Println("Invalid argument error : " + err.Error())
		}
		return nil, fmt.Errorf("error sending data to External Catalog Connector (i = %d): %v", i, err)
	}
	log.Println("***************************************************************")
	log.Printf("Received Response from External Catalog Connector for  dataSetID: %s\n", datasetID)
	log.Println("***************************************************************")
	log.Printf("Response received from External Catalog Connector is given below:")
	responseBytes, errJSON := json.MarshalIndent(r, "", "\t")
	if errJSON != nil {
		log.Printf("error Marshalling Catalog External Connector Response (i = %d): %v", i, errJSON)
		return nil, fmt.Errorf("error Marshalling External Catalog Connector Response (i = %d): %v", i, errJSON)
	}
	log.Print(string(responseBytes))
	log.Println("***************************************************************")
	metadataMap := make(map[string]interface{})
	err = json.Unmarshal(responseBytes, &metadataMap)
	if err != nil {
		log.Printf("error in unmarshalling responseBytes (i = %d): %v", i, err)
		return nil, fmt.Errorf("error in unmarshalling responseBytes (i = %d): %v", i, err)
	}
	inputMap := make(map[string]interface{})
	for k, v := range metadataMap {
		inputMap[k] = v
	}
	return inputMap, nil
}
