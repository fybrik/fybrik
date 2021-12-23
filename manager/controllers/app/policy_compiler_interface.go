// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"log"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	connectors "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
	"fybrik.io/fybrik/pkg/vault"
	"github.com/gdexlab/go-render/render"
)

func ConstructOpenAPIReq(datasetID string, input *app.FybrikApplication, operation *taxonomymodels.PolicyManagerRequestAction) *taxonomymodels.PolicyManagerRequest {
	req := taxonomymodels.PolicyManagerRequest{}
	action := taxonomymodels.PolicyManagerRequestAction{}
	resource := taxonomymodels.Resource{}

	resource.SetName(datasetID)
	req.SetResource(resource)

	action.SetDestination(operation.GetDestination())
	action.SetActionType(operation.GetActionType())
	action.SetProcessingLocation(operation.GetDestination())
	req.SetAction(action)

	reqContext := make(map[string]interface{})
	for k, v := range input.Spec.AppInfo {
		reqContext[k] = v
	}
	req.SetContext(reqContext)

	return &req
}

// LookupPolicyDecisions provides a list of governance actions for the given dataset and the given operation
func LookupPolicyDecisions(datasetID string, policyManager connectors.PolicyManager, input *app.FybrikApplication, op *taxonomymodels.PolicyManagerRequestAction) ([]taxonomymodels.Action, error) {
	// call external policy manager to get governance instructions for this operation
	openapiReq := ConstructOpenAPIReq(datasetID, input, op)
	output := render.AsCode(openapiReq)
	log.Println("constructed openapi request: ", output)

	var creds string
	if input.Spec.SecretRef != "" {
		creds = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}
	openapiResp, err := policyManager.GetPoliciesDecisions(openapiReq, creds)
	var actions []taxonomymodels.Action
	if err != nil {
		return actions, err
	}
	output = render.AsCode(openapiResp)
	log.Println("openapi response received from policy manager: ", output)

	result := openapiResp.GetResult()
	for i := 0; i < len(result); i++ {
		if utils.IsDenied(result[i].GetAction().Name) {
			var message string
			switch *openapiReq.GetAction().ActionType {
			case taxonomymodels.READ:
				message = app.ReadAccessDenied
			case taxonomymodels.WRITE:
				message = app.WriteNotAllowed
			}
			return actions, errors.New(message)
		}
		actions = append(actions, result[i].GetAction())
	}
	return actions, nil
}
