// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"testing"

	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

var EnvValues = map[string]string{
	"CONNECTION_TIMEOUT":    "120",
	"CATALOG_CONNECTOR_URL": "localhost:50084",
	"OPA_SERVER_URL":        "localhost:8282",
}

func GetEnvironment() (int, string, string) {
	timeOutInSecs := EnvValues["CONNECTION_TIMEOUT"]
	timeOutSecs, _ := strconv.Atoi(timeOutInSecs)

	catalogConnectorURL := EnvValues["CATALOG_CONNECTOR_URL"]
	opaServerURL := EnvValues["OPA_SERVER_URL"]
	log.Printf("EnvVariables = %v\n", EnvValues)

	return timeOutSecs, catalogConnectorURL, opaServerURL
}

func GetApplicationContext(purpose string) *pb.ApplicationContext {
	datasetID := "mock-datasetID"
	applicationDetails := &pb.ApplicationDetails{Purpose: purpose, Role: "Security", ProcessingGeography: "US"}
	datasets := []*pb.DatasetContext{}
	datasets = append(datasets, createDatasetRead(datasetID))
	applicationContext := &pb.ApplicationContext{AppInfo: applicationDetails, Datasets: datasets}

	return applicationContext
}

func createDatasetRead(datasetID string) *pb.DatasetContext {
	dataset := &pb.DatasetIdentifier{DatasetId: datasetID}
	operation := &pb.AccessOperation{Type: pb.AccessOperation_READ}
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

func EnsureDeepEqualDecisions(t *testing.T, testedDecisions *pb.PoliciesDecisions, expectedDecisions *pb.PoliciesDecisions) {
	assert.True(t, proto.Equal(testedDecisions, expectedDecisions), "Decisions we got from policyManager are not as expected. Expected: %v, Received: %v", expectedDecisions, testedDecisions)
}

func GetExpectedOpaDecisions(purpose string, in *pb.ApplicationContext) *pb.PoliciesDecisions {
	var dataset = &pb.DatasetIdentifier{DatasetId: "mock-datasetID"}
	var datasetDecison *pb.DatasetDecision
	for _, datasetContext := range in.GetDatasets() {
		operation := datasetContext.GetOperation()
		enforcementActions := make([]*pb.EnforcementAction, 0)
		usedPolicies := make([]*pb.Policy, 0)
		fmt.Println("operation")
		fmt.Println(operation)
		var newUsedPolicy *pb.Policy
		if purpose == "marketing" {
			newEnforcementAction := ConstructEncryptColumn("nameDest")
			enforcementActions = append(enforcementActions, newEnforcementAction)
			newUsedPolicy = &pb.Policy{Description: "test for transactions dataset that encrypts some columns by name"}
			usedPolicies = append(usedPolicies, newUsedPolicy)

			newEnforcementAction = ConstructEncryptColumn("nameOrig")
			enforcementActions = append(enforcementActions, newEnforcementAction)
			newUsedPolicy = &pb.Policy{Description: "test for transactions dataset that encrypts some columns by name"}
			usedPolicies = append(usedPolicies, newUsedPolicy)
		} else {
			newEnforcementAction := ConstructRemoveColumn("nameDest")
			enforcementActions = append(enforcementActions, newEnforcementAction)
			newUsedPolicy = &pb.Policy{Description: "remove columns with name nameOrig and nameDest in datasets which have been tagged with Finance"}
			usedPolicies = append(usedPolicies, newUsedPolicy)

			newEnforcementAction = ConstructRemoveColumn("nameOrig")
			enforcementActions = append(enforcementActions, newEnforcementAction)
			newUsedPolicy = &pb.Policy{Description: "remove columns with name nameOrig and nameDest in datasets which have been tagged with Finance"}
			usedPolicies = append(usedPolicies, newUsedPolicy)
		}
		operationDecision := &pb.OperationDecision{Operation: operation,
			EnforcementActions: enforcementActions,
			UsedPolicies:       usedPolicies}
		datasetDecison = &pb.DatasetDecision{Dataset: dataset, Decisions: []*pb.OperationDecision{operationDecision}}
	}
	return &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison}}
}

// Mocks for catalog-connector and OPA-connector

type connectorMockCatalog struct {
	pb.UnimplementedDataCatalogServiceServer
}

func (s *connectorMockCatalog) GetDatasetInfo(ctx context.Context, req *pb.CatalogDatasetRequest) (*pb.CatalogDatasetInfo, error) {
	return GetCatalogInfo(req.GetAppId(), req.GetDatasetId()), nil
}

func GetCatalogInfo(appID string, datasetID string) *pb.CatalogDatasetInfo {
	var datasetInfo *pb.CatalogDatasetInfo
	componentsMetadata := make(map[string]*pb.DataComponentMetadata)
	componentMetaData1 := &pb.DataComponentMetadata{}
	componentsMetadata["first"] = componentMetaData1
	datasetNamedMetadata := make(map[string]string)

	var datasetDetails *pb.DatasetDetails
	datasetTags := []string{"Tag1"}

	datasetDetails = &pb.DatasetDetails{Name: "mock-name",
		DataOwner: "data-owner",
		DataStore: &pb.DataStore{
			Type: pb.DataStore_LOCAL,
			Name: "mock-name",
		},
		DataFormat: "data-format",
		Geo:        "mock-geo",
		Metadata: &pb.DatasetMetadata{
			DatasetNamedMetadata: datasetNamedMetadata,
			DatasetTags:          datasetTags,
			ComponentsMetadata:   componentsMetadata,
		},
	}

	datasetInfo = &pb.CatalogDatasetInfo{DatasetId: "mock-datasetID", Details: datasetDetails}
	return datasetInfo
}

func MockCatalogConnector(port int) {
	address := utils.ListeningAddress(port)
	log.Println("Start Mock for Catalog Connector at " + address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error in listening: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDataCatalogServiceServer(s, &connectorMockCatalog{})
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}

func customOpaResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: customOpaResponse")

	customeResponse := "{\"result\":{\"deny\":[],\"transform\":[{\"action_name\":\"encrypt column\",\"arguments\":{\"column_name\":\"nameDest\"},\"description\":\"Single column is encrypted with its own key\",\"used_policy\":{\"description\":\"test for transactions dataset that encrypts some columns by name\"}},{\"action_name\":\"encrypt column\",\"arguments\":{\"column_name\":\"nameOrig\"},\"description\":\"Single column is encrypted with its own key\",\"used_policy\":{\"description\":\"test for transactions dataset that encrypts some columns by name\"}}]}}"

	fmt.Fprintf(w, customeResponse)
}

func MockOpaServer(port int) {
	address := utils.ListeningAddress(port)
	log.Println("Start Mock for OPA Server at " + address)

	http.HandleFunc("/v1/data/user_policies", customOpaResponse)
	log.Fatal(http.ListenAndServe(address, nil))
}
