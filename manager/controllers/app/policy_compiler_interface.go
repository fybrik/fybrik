// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	"emperror.dev/errors"
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	connectors "github.com/mesh-for-data/mesh-for-data/pkg/connectors/clients"
	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"github.com/mesh-for-data/mesh-for-data/pkg/vault"
)

// ConstructApplicationContext constructs ApplicationContext structure to send to Policy Compiler
func ConstructApplicationContext(datasetID string, input *app.M4DApplication, operation *pb.AccessOperation) *pb.ApplicationContext {
	var credentialPath string
	if input.Spec.SecretRef != "" {
		credentialPath = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}
	return &pb.ApplicationContext{
		AppInfo: &pb.ApplicationDetails{
			ProcessingGeography: operation.Destination,
			Properties:          input.Spec.AppInfo,
		},
		CredentialPath: credentialPath,
		Datasets: []*pb.DatasetContext{{
			Dataset: &pb.DatasetIdentifier{
				DatasetId: datasetID,
			},
			Operation: operation,
		}},
	}
}

// LookupPolicyDecisions provides a list of governance actions for the given dataset and the given operation
func LookupPolicyDecisions(datasetID string, policyManager connectors.PolicyManager, input *app.M4DApplication, op *pb.AccessOperation) ([]*pb.EnforcementAction, error) {
	// call external policy manager to get governance instructions for this operation
	appContext := ConstructApplicationContext(datasetID, input, op)
	pcresponse, err := policyManager.GetPoliciesDecisions(context.Background(), appContext)
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
