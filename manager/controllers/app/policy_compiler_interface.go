// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"emperror.dev/errors"
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
)

// ConstructApplicationContext constructs ApplicationContext structure to send to Policy Compiler
func ConstructApplicationContext(datasetID string, input *app.M4DApplication, operation *pb.AccessOperation) *pb.ApplicationContext {
	return &pb.ApplicationContext{
		AppInfo: &pb.ApplicationDetails{
			Purpose:             input.Spec.AppInfo.Purpose,
			ProcessingGeography: operation.Destination, //TODO: Remove processing geography, destination is enough
			Role:                string(input.Spec.AppInfo.Role),
		},
		AppId: utils.CreateAppIdentifier(input),
		Datasets: []*pb.DatasetContext{{
			Dataset: &pb.DatasetIdentifier{
				DatasetId: datasetID,
			},
			Operation: operation,
		}},
	}
}

// LookupPolicyDecisions provides a list of governance actions for the given dataset and the given operation
func LookupPolicyDecisions(datasetID string, policyCompiler pc.IPolicyCompiler, input *app.M4DApplication, op *pb.AccessOperation) ([]*pb.EnforcementAction, error) {
	// call external policy manager to get governance instructions for this operation
	appContext := ConstructApplicationContext(datasetID, input, op)
	pcresponse, err := policyCompiler.GetPoliciesDecisions(appContext)
	actions := []*pb.EnforcementAction{}
	if err != nil {
		return actions, err
	}

	for _, datasetDecision := range pcresponse.GetDatasetDecisions() {
		if datasetDecision.GetDataset().GetDatasetId() != datasetID {
			continue // not our data set
		}
		operationDecisions := datasetDecision.GetDecisions()
		for _, operationDecision := range operationDecisions {
			enforcementActions := operationDecision.GetEnforcementActions()
			for _, action := range enforcementActions {
				if utils.IsDenied(action.GetName()) {
					var message string
					switch operationDecision.Operation.Type {
					case pb.AccessOperation_READ:
						message = app.ReadAccessDenied
					case pb.AccessOperation_WRITE:
						message = app.WriteNotAllowed
					}
					return actions, errors.New(message)
				}
				// Check if this is a real action (i.e. not Allow)
				if utils.IsAction(action.GetName()) {
					actions = append(actions, action)
				}
			}
		}
	}
	return actions, nil
}
