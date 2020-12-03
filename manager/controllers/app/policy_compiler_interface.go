// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
)

// ConstructApplicationContext constructs ApplicationContext structure to send to Policy Compiler
func ConstructApplicationContext(datasetID string, input *app.M4DApplication, operationType pb.AccessOperation_AccessType) *pb.ApplicationContext {
	return &pb.ApplicationContext{
		AppInfo: &pb.ApplicationDetails{
			Purpose:             input.Spec.AppInfo.Purpose,
			ProcessingGeography: input.Spec.AppInfo.ProcessingGeography,
			Role:                string(input.Spec.AppInfo.Role),
		},
		AppId: utils.CreateAppIdentifier(input),
		Datasets: []*pb.DatasetContext{{
			Dataset: &pb.DatasetIdentifier{
				DatasetId: datasetID,
			},
			Operation: &pb.AccessOperation{
				Type:        operationType,
				Destination: input.Spec.AppInfo.ProcessingGeography,
			},
		}},
	}
}

// LookupPolicyDecisions provides a list of governance actions for the given dataset and the given operation
func LookupPolicyDecisions(datasetID string, policyCompiler pc.IPolicyCompiler, req *modules.DataInfo, input *app.M4DApplication, op pb.AccessOperation_AccessType) error {
	// call external policy manager to get governance instructions for this operation
	appContext := ConstructApplicationContext(datasetID, input, op)
	var flow app.ModuleFlow
	switch op {
	case pb.AccessOperation_READ:
		flow = app.Read
	case pb.AccessOperation_COPY:
		flow = app.Copy
	case pb.AccessOperation_WRITE:
		flow = app.Write
	}

	pcresponse, err := policyCompiler.GetPoliciesDecisions(appContext)
	if err != nil {
		return err
	}

	// initialize Actions structure
	req.Actions[flow] = modules.Transformations{
		Allowed:            true,
		EnforcementActions: make([]pb.EnforcementAction, 0),
	}

	for _, datasetDecision := range pcresponse.GetDatasetDecisions() {
		if datasetDecision.GetDataset().GetDatasetId() != datasetID {
			continue // not our data set
		}
		var actions []pb.EnforcementAction
		operationDecisions := datasetDecision.GetDecisions()
		for _, operationDecision := range operationDecisions {
			enforcementActions := operationDecision.GetEnforcementActions()
			for _, action := range enforcementActions {
				if utils.IsDenied(action.GetName()) {
					var msg string
					if operationDecision.Operation.Type == pb.AccessOperation_READ {
						msg = app.ReadAccessDenied
					} else {
						msg = app.CopyNotAllowed
					}
					req.Actions[flow] = modules.Transformations{
						Allowed:            false,
						Message:            msg,
						EnforcementActions: make([]pb.EnforcementAction, 0),
					}
					return nil
				}
				// Check if this is a real action (i.e. not Allow)
				if utils.IsAction(action.GetName()) {
					actions = append(actions, *action.DeepCopy())
				}
			}
		}
		req.Actions[flow] = modules.Transformations{
			Allowed:            true,
			EnforcementActions: actions,
		}
	}
	return nil
}
