// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"

	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

var EnvValues = map[string]string{
	"PCSERVER_TIMEOUT":               "120",
	"MAIN_POLICY_MANAGER_NAME":       "WDP",
	"EXTENSIONS_POLICY_MANAGER_NAME": "OPA",
}

func GetEnvironment() (string, string, int, string, string) {
	// mapValues := mapVariableValues()

	timeOutInSecs := EnvValues["PCSERVER_TIMEOUT"]
	timeOutSecs, _ := strconv.Atoi(timeOutInSecs)

	mainPolicyManagerURL := EnvValues["MAIN_POLICY_MANAGER_URL"]
	mainPolicyManagerName := EnvValues["MAIN_POLICY_MANAGER_NAME"]
	extensionPolicyManagerURL := EnvValues["EXTENSIONS_POLICY_MANAGER_URL"]
	extensionPolicyManagerName := EnvValues["EXTENSIONS_POLICY_MANAGER_NAME"]

	log.Printf("EnvVariables = %v\n", EnvValues)

	return mainPolicyManagerName, mainPolicyManagerURL,
		timeOutSecs, extensionPolicyManagerName, extensionPolicyManagerURL
}

func GetApplicationContext(purpose string) *pb.ApplicationContext {
	datasetID := "mock-datasetID"
	applicationDetails := &pb.ApplicationDetails{Properties: map[string]string{"intent": purpose, "role": "Security"}, ProcessingGeography: "US"}
	datasets := []*pb.DatasetContext{}
	datasets = append(datasets, createDatasetRead(datasetID))
	datasets = append(datasets, createDatasetTransferFirst(datasetID))
	applicationContext := &pb.ApplicationContext{AppInfo: applicationDetails, Datasets: datasets}

	return applicationContext
}

func createDatasetRead(datasetID string) *pb.DatasetContext {
	dataset := &pb.DatasetIdentifier{DatasetId: datasetID}
	operation := &pb.AccessOperation{Type: pb.AccessOperation_READ}
	datasetContext := &pb.DatasetContext{Dataset: dataset, Operation: operation}
	return datasetContext
}

func createDatasetTransferFirst(datasetID string) *pb.DatasetContext {
	dataset := &pb.DatasetIdentifier{DatasetId: datasetID}
	operation := &pb.AccessOperation{Type: pb.AccessOperation_COPY, Destination: "US"}
	datasetContext := &pb.DatasetContext{Dataset: dataset, Operation: operation}
	return datasetContext
}

func ConstructRemoveColumn(colName string) *pb.EnforcementAction {
	return &pb.EnforcementAction{Name: "removed", Id: "removed-ID",
		Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": colName}}
}

func ConstructEncryptColumn(colName string) *pb.EnforcementAction {
	return &pb.EnforcementAction{Name: "encrypted", Id: "encrypted-ID",
		Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": colName}}
}

func ConstructRedactColumn(colName string) *pb.EnforcementAction {
	return &pb.EnforcementAction{Name: "redact", Id: "redact-ID",
		Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": colName}}
}

func ConstructAllow() *pb.EnforcementAction {
	return &pb.EnforcementAction{Name: "Allow", Id: "Allow-ID",
		Level: pb.EnforcementAction_DATASET, Args: map[string]string{}}
}

func EnsureDeepEqualDecisions(t *testing.T, testedDecisions *pb.PoliciesDecisions, expectedDecisions *pb.PoliciesDecisions) {
	assert.True(t, proto.Equal(testedDecisions, expectedDecisions), "Decisions we got from policyManager are not as expected. Expected: %v, Received: %v", expectedDecisions, testedDecisions)
}

func VerifyContainsDatasetDecision(t *testing.T, combinedPolicies *pb.PoliciesDecisions, datasetDecison *pb.DatasetDecision) {
	for _, combinedDatasetPolicies := range combinedPolicies.DatasetDecisions {
		if proto.Equal(combinedDatasetPolicies.Dataset, datasetDecison.Dataset) {
			for _, operationDecision := range datasetDecison.Decisions {
				VerifyContainsSingleOperationDesision(t, combinedDatasetPolicies, operationDecision)
			}
			// found correct dataset and all operation-desioins inside it
			return
		}
	}
	assert.Fail(t, "didn't find the correct deision for this dataset", datasetDecison.Dataset)
}

func VerifyContainsSingleOperationDesision(t *testing.T, combinedDatasetPolicies *pb.DatasetDecision, operationDesicion *pb.OperationDecision) {
	for _, decision := range combinedDatasetPolicies.Decisions {
		if proto.Equal(decision.Operation, operationDesicion.Operation) {
			// found correct decision for this dataset and access operation
			for _, action := range operationDesicion.EnforcementActions {
				isFound := false
				for _, actionCombined := range decision.EnforcementActions {
					if proto.Equal(action, actionCombined) {
						isFound = true
						break
					}
				}
				if !isFound {
					// one of the actions in dataset action is not in combined decisions
					assert.Fail(t, "combined desisions miss enforcement action ", action)
				}
			}
			for _, usedPolicy := range operationDesicion.UsedPolicies {
				isFound := false
				for _, policyCombined := range decision.UsedPolicies {
					if proto.Equal(usedPolicy, policyCombined) {
						isFound = true
						break
					}
				}
				if !isFound {
					// one of the usedPolicies in dataset action is not in combined decisions
					assert.Fail(t, "combined desisions miss used policy ", usedPolicy)
				}
			}
			return // found correct desision and it contains all enforceemnt actions and used policies
		}
	}
	assert.Fail(t, "didn't find the correct deision for this operation ", operationDesicion.Operation)
}

func CheckPolicies(t *testing.T, policies *pb.PoliciesDecisions, decision1, decision2 *pb.PoliciesDecisions) {
	for _, datasetDecisions := range decision1.DatasetDecisions {
		VerifyContainsDatasetDecision(t, policies, datasetDecisions)
	}
	for _, datasetDecisions := range decision2.DatasetDecisions {
		VerifyContainsDatasetDecision(t, policies, datasetDecisions)
	}
}

/****************************/
// Connector Mock "Main policy manager"

type connectorMockMain struct {
	pb.UnimplementedPolicyManagerServiceServer
}

func GetMainPMDecisions(purpose string) *pb.PoliciesDecisions {
	var dataset = &pb.DatasetIdentifier{DatasetId: "mock-datasetID"}
	var column = "mock-col-1"
	var usedPolicy = &pb.Policy{Description: "policy 1 description"}
	var datasetDecison *pb.DatasetDecision

	if purpose == "fraud-detection" {
		operationDecision := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
			EnforcementActions: []*pb.EnforcementAction{ConstructRemoveColumn(column)},
			UsedPolicies:       []*pb.Policy{usedPolicy}}
		datasetDecison = &pb.DatasetDecision{Dataset: dataset, Decisions: []*pb.OperationDecision{operationDecision}}
	} else {
		operationDecision := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
			EnforcementActions: []*pb.EnforcementAction{ConstructRedactColumn(column)},
			UsedPolicies:       []*pb.Policy{usedPolicy}}
		datasetDecison = &pb.DatasetDecision{Dataset: dataset, Decisions: []*pb.OperationDecision{operationDecision}}
	}

	return &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison}}
}

func (s *connectorMockMain) GetPoliciesDecisions(ctx context.Context, in *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
	fmt.Println("Output for applicationContext")
	fmt.Println(in.AppInfo.Properties["intent"])

	return GetMainPMDecisions(in.AppInfo.Properties["intent"]), nil
}

func MockMainConnector(port int) {
	address := utils.ListeningAddress(port)
	log.Println("Start Mock for Main PolicyConnector at " + address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error in listening: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPolicyManagerServiceServer(s, &connectorMockMain{})
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}

/****************************/
// Connector Mock "Extension policy manager"

type connectorMockExt struct {
	pb.UnimplementedPolicyManagerServiceServer
}

func GetExtPMDecisions(purpose string) *pb.PoliciesDecisions {
	var dataset = &pb.DatasetIdentifier{DatasetId: "mock-datasetID"}
	var column = "mock-col-2"
	var usedPolicy = &pb.Policy{Description: "policy 2 description"}
	var datasetDecison *pb.DatasetDecision

	if purpose == "fraud-detection" {
		operationDecision := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
			EnforcementActions: []*pb.EnforcementAction{ConstructRedactColumn(column)},
			UsedPolicies:       []*pb.Policy{usedPolicy}}
		datasetDecison = &pb.DatasetDecision{Dataset: dataset, Decisions: []*pb.OperationDecision{operationDecision}}
	} else {
		operationDecision := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
			EnforcementActions: []*pb.EnforcementAction{ConstructEncryptColumn(column), ConstructRemoveColumn(column)},
			UsedPolicies:       []*pb.Policy{usedPolicy}}
		datasetDecison = &pb.DatasetDecision{Dataset: dataset, Decisions: []*pb.OperationDecision{operationDecision}}
	}

	return &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison}}
}

func (s *connectorMockExt) GetPoliciesDecisions(ctx context.Context, in *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
	return GetExtPMDecisions(in.AppInfo.Properties["intent"]), nil
}

func MockExtConnector(port int) {
	address := utils.ListeningAddress(port)
	log.Println("Start Mock for Extension PolicyConnector at port " + address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error in listening: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPolicyManagerServiceServer(s, &connectorMockExt{})
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}
