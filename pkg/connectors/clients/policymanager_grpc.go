// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"emperror.dev/errors"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
	"google.golang.org/grpc"
)

var _ PolicyManager = (*grpcPolicyManager)(nil)

type grpcPolicyManager struct {
	pb.UnimplementedPolicyManagerServiceServer

	name       string
	connection *grpc.ClientConn
	client     pb.PolicyManagerServiceClient
}

// NewGrpcPolicyManager creates a PolicyManager facade that connects to a GRPC service
// You must call .Close() when you are done using the created instance
func NewGrpcPolicyManager(name string, connectionURL string, connectionTimeout time.Duration) (PolicyManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	connection, err := grpc.DialContext(ctx, connectionURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("NewGrpcPolicyManager failed when connecting to %s", connectionURL))
	}
	return &grpcPolicyManager{
		name:       name,
		client:     pb.NewPolicyManagerServiceClient(connection),
		connection: connection,
	}, nil
}

func (m *grpcPolicyManager) GetPoliciesDecisions(
	in *openapiclientmodels.PolicyManagerRequest, creds string) (*openapiclientmodels.PolicyManagerResponse, error) {

	appContext, _ := ConvertOpenApiReqToGrpcReq(in, creds)
	log.Println("appContext: created from convertOpenApiReqToGrpcReq: ", appContext)

	result, _ := m.client.GetPoliciesDecisions(context.Background(), appContext)

	log.Println("GRPC result returned from GetPoliciesDecisions:", result)
	policyManagerResp, _ := ConvertGrpcRespToOpenApiResp(result)

	res, err := json.MarshalIndent(policyManagerResp, "", "\t")
	log.Println("err :", err)
	log.Println("policyManagerResp: created from convGrpcRespToOpenApiResp")
	log.Println("marshalled response:", string(res))
	return policyManagerResp, nil

	// result, err := m.client.GetPoliciesDecisions(ctx, in)
	// return result, errors.Wrap(err, fmt.Sprintf("get policies decisions from %s failed", m.name))
}

func (m *grpcPolicyManager) Close() error {
	return m.connection.Close()
}

// ref: https://sosedoff.com/2014/12/15/generate-random-hex-string-in-go.html
func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ConvertGrpcReqToOpenApiReq(in *pb.ApplicationContext) (*openapiclientmodels.PolicyManagerRequest, string, error) {

	req := openapiclientmodels.PolicyManagerRequest{}
	action := openapiclientmodels.PolicyManagerRequestAction{}
	resource := openapiclientmodels.Resource{}

	creds := in.GetCredentialPath()

	datasets := in.GetDatasets()
	// assume only one dataset is passed
	for i := 0; i < len(datasets); i++ {
		operation := datasets[i].GetOperation()
		operationType := operation.GetType()
		if operationType == pb.AccessOperation_READ {
			action.SetActionType(openapiclientmodels.READ)
		}
		if operationType == pb.AccessOperation_WRITE {
			action.SetActionType(openapiclientmodels.WRITE)
		}
		datasetID := datasets[i].GetDataset().GetDatasetId()
		resource.SetName(datasetID)
	}
	req.SetAction(action)
	req.SetResource(resource)

	processingGeo := in.GetAppInfo().GetProcessingGeography()
	action.SetProcessingLocation(processingGeo)

	reqContext := make(map[string]interface{})
	properties := in.GetAppInfo().GetProperties()
	reqContext["intent"] = properties["intent"]
	reqContext["role"] = properties["role"]
	req.SetContext(reqContext)

	return &req, creds, nil

}

func ConvertOpenApiReqToGrpcReq(in *openapiclientmodels.PolicyManagerRequest, creds string) (*pb.ApplicationContext, error) {

	credentialPath := creds
	action := in.GetAction()
	processingGeo := (&action).GetProcessingLocation()
	log.Println("processingGeo: ", processingGeo)

	properties := make(map[string]string)
	context := in.GetContext()
	if intent, ok := context["intent"].(string); ok {
		properties["intent"] = intent
	}
	if role, ok := context["role"].(string); ok {
		properties["role"] = role
	}

	appInfo := &pb.ApplicationDetails{ProcessingGeography: string(processingGeo), Properties: properties}

	datasetContextList := []*pb.DatasetContext{}
	resource := in.GetResource()
	datasetId := (&resource).GetName()
	dataset := &pb.DatasetIdentifier{DatasetId: datasetId}
	// ?? this is not supported in openapi
	destination := ""
	actionType := (&action).GetActionType()

	var grpcActionType pb.AccessOperation_AccessType
	if actionType == openapiclientmodels.READ {
		grpcActionType = pb.AccessOperation_READ
	} else if actionType == openapiclientmodels.WRITE {
		grpcActionType = pb.AccessOperation_WRITE
	} else {
		// default is read
		grpcActionType = pb.AccessOperation_READ
	}

	operation := &pb.AccessOperation{Type: grpcActionType, Destination: destination}
	datasetContext := &pb.DatasetContext{Dataset: dataset, Operation: operation}
	datasetContextList = append(datasetContextList, datasetContext)

	appContext := &pb.ApplicationContext{CredentialPath: credentialPath, AppInfo: appInfo, Datasets: datasetContextList}

	log.Println("Constructed GRPC appContext: ", appContext)

	return appContext, nil
}

func ConvertOpenApiRespToGrpcResp(
	out *openapiclientmodels.PolicyManagerResponse,
	datasetID string, op *pb.AccessOperation) (*pb.PoliciesDecisions, error) {

	resultItems := out.GetResult()
	enforcementActions := make([]*pb.EnforcementAction, 0)
	usedPolicies := make([]*pb.Policy, 0)

	for i := 0; i < len(resultItems); i++ {
		action := resultItems[i].GetAction()
		name := action.GetName()
		additionalProperties := action.AdditionalProperties

		if name == "redact" {
			if additionalProperties != nil {
				if colNames, ok := additionalProperties["columns"].([]string); ok {
					for j := 0; j < len(colNames); j++ {
						newEnforcementAction := &pb.EnforcementAction{Name: "redact", Id: "redact-ID",
							Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": colNames[j]}}
						enforcementActions = append(enforcementActions, newEnforcementAction)

						policy := resultItems[i].GetPolicy()
						newUsedPolicy := &pb.Policy{Description: policy}
						usedPolicies = append(usedPolicies, newUsedPolicy)
					}
				}
			}
		}

		if name == "remove" {
			if additionalProperties != nil {
				if colNames, ok := additionalProperties["columns"].([]string); ok {
					for j := 0; j < len(colNames); j++ {
						newEnforcementAction := &pb.EnforcementAction{Name: "removed", Id: "removed-ID",
							Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": colNames[j]}}
						enforcementActions = append(enforcementActions, newEnforcementAction)

						policy := resultItems[i].GetPolicy()
						newUsedPolicy := &pb.Policy{Description: policy}
						usedPolicies = append(usedPolicies, newUsedPolicy)
					}
				}
			}
		}

		if name == "encrypt" {
			if additionalProperties != nil {
				if colNames, ok := additionalProperties["columns"].([]string); ok {
					for j := 0; j < len(colNames); j++ {
						newEnforcementAction := &pb.EnforcementAction{Name: "encrypted", Id: "encrypted-ID",
							Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": colNames[j]}}
						enforcementActions = append(enforcementActions, newEnforcementAction)

						policy := resultItems[i].GetPolicy()
						newUsedPolicy := &pb.Policy{Description: policy}
						usedPolicies = append(usedPolicies, newUsedPolicy)
					}
				}
			}
		}

		if name == "deny" {
			newEnforcementAction := &pb.EnforcementAction{Name: "Deny", Id: "Deny-ID", Level: pb.EnforcementAction_DATASET, Args: map[string]string{}}
			enforcementActions = append(enforcementActions, newEnforcementAction)

			policy := resultItems[i].GetPolicy()
			newUsedPolicy := &pb.Policy{Description: policy}
			usedPolicies = append(usedPolicies, newUsedPolicy)
		}
	}

	opaOperationDecision := &pb.OperationDecision{Operation: op, EnforcementActions: enforcementActions, UsedPolicies: usedPolicies}

	var datasetDecisionList []*pb.DatasetDecision
	var opaOperationDecisionList []*pb.OperationDecision
	opaOperationDecisionList = append(opaOperationDecisionList, opaOperationDecision)
	// Create a new *DatasetDecision
	dataset := &pb.DatasetIdentifier{DatasetId: datasetID}
	datasetDecison := &pb.DatasetDecision{Dataset: dataset, Decisions: opaOperationDecisionList}
	datasetDecisionList = append(datasetDecisionList, datasetDecison)

	return &pb.PoliciesDecisions{DatasetDecisions: datasetDecisionList}, nil
}

func ConvertGrpcRespToOpenApiResp(result *pb.PoliciesDecisions) (*openapiclientmodels.PolicyManagerResponse, error) {

	// convert GRPC response to Open Api Response - start
	//decisionId := "3ffb47c7-c3c7-4fe7-b244-b38dc8951b87"
	// we dont get decision id returned from OPA from GRPC response. So we generate random hex string
	decisionId, _ := randomHex(20)
	log.Println("decision id generated", decisionId)

	var datasetDecisions []*pb.DatasetDecision
	var decisions []*pb.OperationDecision
	datasetDecisions = result.GetDatasetDecisions()
	respResult := []openapiclientmodels.ResultItem{}

	// we assume only one dataset decision is passed
	for i := 0; i < len(datasetDecisions); i++ {
		datasetDecision := datasetDecisions[i]
		//datasetID := datasetDecision.GetDataset()
		decisions = datasetDecision.GetDecisions()

		for j := 0; j < len(decisions); j++ {
			decision := decisions[j]
			//operation := decision.GetOperation()

			var enfActionList []*pb.EnforcementAction
			var usedPoliciesList []*pb.Policy
			enfActionList = decision.GetEnforcementActions()
			usedPoliciesList = decision.GetUsedPolicies()

			for k := 0; k < len(enfActionList); k++ {
				enfAction := enfActionList[k]
				name := enfAction.GetName()
				level := enfAction.GetLevel()
				args := enfAction.GetArgs()
				log.Println("args received: ", args)
				log.Println("name received: ", name)
				policyManagerResult := openapiclientmodels.ResultItem{}

				if level == pb.EnforcementAction_COLUMN {
					actionOnCols := openapiclientmodels.Action{}
					action := make(map[string]interface{})
					if name == "redact" {
						action["name"] = "redact"
						action["columns"] = []string{args["column_name"]}
					}
					if name == "encrypt" {
						action["name"] = "encrypt"
						action["columns"] = []string{args["column_name"]}
					}
					if name == "remove" {
						action["name"] = "remove"
						action["columns"] = []string{args["column_name"]}
					}

					actionBytes, errJSON := json.MarshalIndent(action, "", "\t")
					if errJSON != nil {
						return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
					}
					log.Println("actionBytes:", string(actionBytes))
					err := json.Unmarshal(actionBytes, &actionOnCols)
					if err != nil {
						return nil, fmt.Errorf("error in unmarshalling actionBytes : %v", err)
					}
					// just for printing
					actionOnColsBytes, errJSON := json.MarshalIndent(actionOnCols, "", "\t")
					if errJSON != nil {
						return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
					}
					log.Println("actionOnColsBytes: ", string(actionOnColsBytes))

					policyManagerResult.SetAction(actionOnCols)
				}

				if level == pb.EnforcementAction_DATASET {
					actionOnDataset := openapiclientmodels.Action{}
					if name == "Deny" {
						actionOnDataset.SetName("deny")
					}
					policyManagerResult.SetAction(actionOnDataset)
				}
				if k < len(usedPoliciesList) {
					policy := usedPoliciesList[k].GetDescription()
					log.Println("usedPoliciesList[k].GetDescription()", policy)
					policyManagerResult.SetPolicy(policy)
				}
				respResult = append(respResult, policyManagerResult)
			}
		}
	}
	// convert GRPC response to Open Api Response - end
	policyManagerResp := &openapiclientmodels.PolicyManagerResponse{DecisionId: &decisionId, Result: respResult}

	log.Println("policyManagerResp in convGrpcRespToOpenApiResp", policyManagerResp)

	return policyManagerResp, nil
}
