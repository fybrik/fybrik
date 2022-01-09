// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"log"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	connectors "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	"fybrik.io/fybrik/pkg/model/policymanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/vault"
	"github.com/gdexlab/go-render/render"
)

func ConstructOpenAPIReq(datasetID string, input *app.FybrikApplication, operation *policymanager.RequestAction) (*policymanager.GetPolicyDecisionsRequest, error) {
	context, err := utils.StructToMap(&input.Spec.AppInfo)
	if err != nil {
		return nil, err
	}
	return &policymanager.GetPolicyDecisionsRequest{
		Context: taxonomy.PolicyManagerRequestContext{Properties: serde.Properties{
			Items: context,
		}},
		Action: *operation,
		Resource: policymanager.Resource{
			ID: taxonomy.AssetID(datasetID),
		},
	}, nil
}

// LookupPolicyDecisions provides a list of governance actions for the given dataset and the given operation
func LookupPolicyDecisions(datasetID string, policyManager connectors.PolicyManager, input *app.FybrikApplication, op *policymanager.RequestAction) ([]taxonomy.Action, error) {
	// call external policy manager to get governance instructions for this operation
	openapiReq, err := ConstructOpenAPIReq(datasetID, input, op)
	if err != nil {
		return nil, err
	}
	output := render.AsCode(openapiReq)
	log.Println("constructed openapi request: ", output)

	var creds string
	if input.Spec.SecretRef != "" {
		creds = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}
	openapiResp, err := policyManager.GetPoliciesDecisions(openapiReq, creds)
	var actions []taxonomy.Action
	if err != nil {
		return actions, err
	}
	output = render.AsCode(openapiResp)
	log.Println("openapi response received from policy manager: ", output)

	result := openapiResp.Result
	for i := 0; i < len(result); i++ {
		if utils.IsDenied(result[i].Action.Name) {
			var message string
			switch openapiReq.Action.ActionType {
			case policymanager.READ:
				message = app.ReadAccessDenied
			case policymanager.WRITE:
				message = app.WriteNotAllowed
			}
			return actions, errors.New(message)
		}
		actions = append(actions, result[i].Action)
	}
	return actions, nil
}
