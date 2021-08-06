// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	connectors "fybrik.io/fybrik/pkg/connectors/clients"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
)

// MockPolicyManager is a mock for PolicyManager interface used in tests
type MockPolicyManager struct {
	connectors.PolicyManager
}

// GetPoliciesDecisions implements the PolicyCompiler interface
// func (s *MockPolicyManager) GetPoliciesDecisions(ctx context.Context, in *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
func (m *MockPolicyManager) GetPoliciesDecisions(
	input *openapiclientmodels.PolicyManagerRequest, creds string) (*openapiclientmodels.PolicyManagerResponse, error) {

	in, _ := connectors.ConvertOpenApiReqToGrpcReq(input, creds)
	log.Println("appContext: created from convertOpenApiReqToGrpcReq: ", in)

	log.Printf("Received: ")
	log.Printf("ProcessingGeography: " + in.AppInfo.GetProcessingGeography())
	log.Printf("Secret: " + in.GetCredentialPath())
	log.Printf("Properties:")
	for key, val := range in.AppInfo.GetProperties() {
		log.Printf(key + " : " + val)
	}
	var externalComponents []*pb.ComponentVersion
	externalComponents = append(externalComponents, &pb.ComponentVersion{Id: "PC1", Version: "1.0", Name: "PolicyCompiler"})
	var dataSetWithActions []*pb.DatasetDecision

	for ind, element := range in.GetDatasets() {
		dataset := element.GetDataset()
		log.Printf("Sending DataSet: ")
		log.Printf("   DataSetID: " + dataset.GetDatasetId())
		var enforcementActions []*pb.EnforcementAction
		args := make(map[string]string)

		var operationDecisions []*pb.OperationDecision
		splittedID := strings.SplitN(dataset.GetDatasetId(), "/", 2)
		if len(splittedID) != 2 {
			panic(fmt.Sprintf("Invalid dataset ID for mock: %s", dataset.GetDatasetId()))
		}
		assetID := splittedID[1]
		switch assetID {
		case "allow-dataset":
			enforcementActions = append(enforcementActions, &pb.EnforcementAction{
				Name: "Allow",
				Id:   "Allow-ID",
			})
		case "deny-dataset":
			enforcementActions = append(enforcementActions, &pb.EnforcementAction{
				Name: "Deny",
				Id:   "Deny-ID",
			})
		case "allow-theshire":
			if element.GetOperation().Destination == "theshire" {
				enforcementActions = append(enforcementActions, &pb.EnforcementAction{
					Name: "Allow",
					Id:   "Allow-ID",
				})
			} else {
				enforcementActions = append(enforcementActions, &pb.EnforcementAction{
					Name: "Deny",
					Id:   "Deny-ID",
				})
			}
		case "deny-theshire":
			if element.GetOperation().Destination != "theshire" {
				enforcementActions = append(enforcementActions, &pb.EnforcementAction{
					Name: "Allow",
					Id:   "Allow-ID",
				})
			} else {
				enforcementActions = append(enforcementActions, &pb.EnforcementAction{
					Name: "Deny",
					Id:   "Deny-ID",
				})
			}
		default:
			args["column"] = "SSN"
			enforcementActions = append(enforcementActions, &pb.EnforcementAction{
				Name:  "redact",
				Id:    "redact-ID",
				Level: pb.EnforcementAction_COLUMN,
				Args:  args})
		}
		operationDecisions = append(operationDecisions, &pb.OperationDecision{Operation: in.GetDatasets()[0].GetOperation(), EnforcementActions: enforcementActions})
		dataSetWithActions = append(dataSetWithActions, &pb.DatasetDecision{
			Dataset: &pb.DatasetIdentifier{
				DatasetId: in.GetDatasets()[ind].GetDataset().GetDatasetId()},
			Decisions: operationDecisions})
	}

	result := &pb.PoliciesDecisions{ComponentVersions: externalComponents,
		DatasetDecisions: dataSetWithActions}

	policyManagerResp, _ := connectors.ConvertGrpcRespToOpenApiResp(result)

	res, err := json.MarshalIndent(policyManagerResp, "", "\t")
	log.Println("err :", err)
	log.Println("policyManagerResp: created from convGrpcRespToOpenApiResp")
	log.Println("marshalled response:", string(res))

	return policyManagerResp, nil
}
