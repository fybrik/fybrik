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
	// mapValues := mapVariableValues()

	timeOutInSecs := EnvValues["CONNECTION_TIMEOUT"]
	timeOutSecs, _ := strconv.Atoi(timeOutInSecs)

	catalogConnectorURL := EnvValues["CATALOG_CONNECTOR_URL"]
	opaServerUrl := EnvValues["OPA_SERVER_URL"]
	log.Printf("EnvVariables = %v\n", EnvValues)

	return timeOutSecs, catalogConnectorURL, opaServerUrl
}

func GetApplicationContext(purpose string) *pb.ApplicationContext {
	datasetID := "mock-datasetID"
	applicationDetails := &pb.ApplicationDetails{Purpose: purpose, Role: "Security", ProcessingGeography: "US"}
	datasets := []*pb.DatasetContext{}
	datasets = append(datasets, createDatasetRead(datasetID))
	//datasets = append(datasets, createDatasetTransferFirst(datasetID))
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

/*func ConstructEncryptColumn(colName string) *pb.EnforcementAction {
	return &pb.EnforcementAction{Name: "encrypted", Id: "encrypted-ID",
		Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": colName}}
}*/

func ConstructRedactColumn(colName string) *pb.EnforcementAction {
	return &pb.EnforcementAction{Name: "redact", Id: "redact-ID",
		Level: pb.EnforcementAction_COLUMN, Args: map[string]string{"column_name": colName}}
}

/*func ConstructAllow() *pb.EnforcementAction {
	return &pb.EnforcementAction{Name: "Allow", Id: "Allow-ID",
		Level: pb.EnforcementAction_DATASET, Args: map[string]string{}}
}*/

func EnsureDeepEqualDecisions(t *testing.T, testedDecisions *pb.PoliciesDecisions, expectedDecisions *pb.PoliciesDecisions) {
	assert.True(t, proto.Equal(testedDecisions, expectedDecisions), "Decisions we got from policyManager are not as expected. Expected: %v, Received: %v", expectedDecisions, testedDecisions)
}

func VerifyContainsDatasetDecision(t *testing.T, combinedPolicies *pb.PoliciesDecisions, datasetDecison *pb.DatasetDecision) {
	for _, combinedDatasetPolicies := range combinedPolicies.DatasetDecisions {
		if proto.Equal(combinedDatasetPolicies.Dataset, datasetDecison.Dataset) {
			for _, operationDecision := range datasetDecison.Decisions {
				VerifyContainsSingleOperationDesision(t, combinedDatasetPolicies, operationDecision)
			}
			//found correct dataset and all operation-desioins inside it
			return
		}
	}
	assert.Fail(t, "didn't find the correct deision for this dataset", datasetDecison.Dataset)
}

func VerifyContainsSingleOperationDesision(t *testing.T, combinedDatasetPolicies *pb.DatasetDecision, operationDesicion *pb.OperationDecision) {
	for _, decision := range combinedDatasetPolicies.Decisions {
		if proto.Equal(decision.Operation, operationDesicion.Operation) {
			//found correct decision for this dataset and access operation
			for _, action := range operationDesicion.EnforcementActions {
				isFound := false
				for _, actionCombined := range decision.EnforcementActions {
					if proto.Equal(action, actionCombined) {
						isFound = true
						break
					}
				}
				if !isFound {
					//one of the actions in dataset action is not in combined decisions
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
					//one of the usedPolicies in dataset action is not in combined decisions
					assert.Fail(t, "combined desisions miss used policy ", usedPolicy)
				}
			}
			return //found correct desision and it contains all enforceemnt actions and used policies
		}
	}
	assert.Fail(t, "didn't find the correct deision for this operation ", operationDesicion.Operation)
}

/*func CheckPolicies(t *testing.T, policies *pb.PoliciesDecisions, decision1, decision2 *pb.PoliciesDecisions) {
	for _, datasetDecisions := range decision1.DatasetDecisions {
		VerifyContainsDatasetDecision(t, policies, datasetDecisions)
	}
	for _, datasetDecisions := range decision2.DatasetDecisions {
		VerifyContainsDatasetDecision(t, policies, datasetDecisions)
	}
}*/

func GetExpectedOpaDecisions(purpose string, in *pb.ApplicationContext) *pb.PoliciesDecisions {
	var dataset = &pb.DatasetIdentifier{DatasetId: "mock-datasetID"}
	var datasetDecison *pb.DatasetDecision
	for _, datasetContext := range in.GetDatasets() {
		operation := datasetContext.GetOperation()
		enforcementActions := make([]*pb.EnforcementAction, 0)
		usedPolicies := make([]*pb.Policy, 0)
		operation = datasetContext.GetOperation()
		fmt.Println("operation")
		fmt.Println(operation)
		var newUsedPolicy *pb.Policy
		if purpose == "marketing" {
			newEnforcementAction := ConstructRedactColumn("nameDest")
			enforcementActions = append(enforcementActions, newEnforcementAction)
			newUsedPolicy = &pb.Policy{Description: "reduct columns with name nameOrig and nameDest  in datasets with Finance"}
			usedPolicies = append(usedPolicies, newUsedPolicy)

			newEnforcementAction = ConstructRedactColumn("nameOrig")
			enforcementActions = append(enforcementActions, newEnforcementAction)
			newUsedPolicy = &pb.Policy{Description: "reduct columns with name nameOrig and nameDest  in datasets with Finance"}
			usedPolicies = append(usedPolicies, newUsedPolicy)
			// newUsedPolicy = &pb.Policy{Description: "reduct columns with name nameOrig and nameDest  in datasets with Finance"}
		} else {
			newEnforcementAction := ConstructRemoveColumn("nameDest")
			enforcementActions = append(enforcementActions, newEnforcementAction)
			newUsedPolicy = &pb.Policy{Description: "remove columns with name nameOrig and nameDest  in datasets with Finance"}
			usedPolicies = append(usedPolicies, newUsedPolicy)

			newEnforcementAction = ConstructRemoveColumn("nameOrig")
			enforcementActions = append(enforcementActions, newEnforcementAction)
			newUsedPolicy = &pb.Policy{Description: "remove columns with name nameOrig and nameDest  in datasets with Finance"}
			usedPolicies = append(usedPolicies, newUsedPolicy)
			// newUsedPolicy = &pb.Policy{Description: "remove columns with name nameOrig and nameDest  in datasets with Finance"}
		}
		//usedPolicies = append(usedPolicies, newUsedPolicy)
		operationDecision := &pb.OperationDecision{Operation: operation,
			EnforcementActions: enforcementActions,
			UsedPolicies:       usedPolicies}
		datasetDecison = &pb.DatasetDecision{Dataset: dataset, Decisions: []*pb.OperationDecision{operationDecision}}
	}
	return &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison}}
}

/****************************/
//Connector Mock "Extension policy manager"

type connectorMockCatalog struct {
	pb.UnimplementedDataCatalogServiceServer
}

func GetCatalogInfo(appId string, datasetID string) *pb.CatalogDatasetInfo {
	var datasetInfo *pb.CatalogDatasetInfo
	var componentsMetadata map[string]*pb.DataComponentMetadata
	componentsMetadata = make(map[string]*pb.DataComponentMetadata)
	componentMetaData1 := &pb.DataComponentMetadata{}
	componentsMetadata["first"] = componentMetaData1
	var datasetNamedMetadata map[string]string
	datasetNamedMetadata = make(map[string]string)

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

func (s *connectorMockCatalog) GetDatasetInfo(ctx context.Context, req *pb.CatalogDatasetRequest) (*pb.CatalogDatasetInfo, error) {
	return GetCatalogInfo(req.GetAppId(), req.GetDatasetId()), nil
}

func MockCatalogConnector(port string) {
	log.Println("Start Mock for Catalog Connector at port " + port)

	lis, err := net.Listen("tcp", ":"+port)
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

	customeResponse := "{\"result\":{\"allow\":true,\"allowed_access_types\":[\"READ\",\"COPY\",\"WRITE\"],\"allowed_copy_destinations\":[\"NorthAmerica\",\"US\"],\"allowed_purposes\":[\"analysis\",\"fraud-detection\"],\"allowed_roles\":[\"DataScientist\",\"Security\"],\"deny\":[],\"transform\":[{\"args\":{\"column name\":\"nameDest\"},\"result\":\"Redact column\",\"used_policy\":{\"description\":\"reduct columns with name nameOrig and nameDest  in datasets with Finance\"}},{\"args\":{\"column name\":\"nameOrig\"},\"result\":\"Redact column\",\"used_policy\":{\"description\":\"reduct columns with name nameOrig and nameDest  in datasets with Finance\"}}]}}"

	fmt.Fprintf(w, customeResponse)
}

func MockOpaServer(port string) {
	log.Println("Start Mock for OPA Server at port " + port)

	http.HandleFunc("/v1/data/extendedEnforcement", customOpaResponse)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
