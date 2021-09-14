// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"log"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	connectors "fybrik.io/fybrik/pkg/connectors/clients"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
	"fybrik.io/fybrik/pkg/vault"
)

func ConstructOpenAPIReq(datasetID string, input *app.FybrikApplication, operation *openapiclientmodels.PolicyManagerRequestAction) (*openapiclientmodels.PolicyManagerRequest, string, error) {
	req := openapiclientmodels.PolicyManagerRequest{}
	action := openapiclientmodels.PolicyManagerRequestAction{}
	resource := openapiclientmodels.Resource{}

	resource.SetName(datasetID)
	req.SetResource(resource)

	destination := operation.GetDestination()
	action.SetDestination(destination)
	operationType := operation.GetActionType()
	if operationType == openapiclientmodels.READ {
		action.SetActionType(openapiclientmodels.READ)
	}
	if operationType == openapiclientmodels.WRITE {
		action.SetActionType(openapiclientmodels.WRITE)
	}
	action.SetProcessingLocation(operation.GetDestination())
	req.SetAction(action)

	reqContext := make(map[string]interface{})
	for k, v := range input.Spec.AppInfo {
		reqContext[k] = v
	}
	req.SetContext(reqContext)

	var credentialPath string
	if input.Spec.SecretRef != "" {
		credentialPath = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}

	return &req, credentialPath, nil
}

// LookupPolicyDecisions provides a list of governance actions for the given dataset and the given operation
func LookupPolicyDecisions(datasetID string, policyManager connectors.PolicyManager, input *app.FybrikApplication, op *openapiclientmodels.PolicyManagerRequestAction) ([]*openapiclientmodels.ResultItem, error) {
	// call external policy manager to get governance instructions for this operation
	openapiReq, creds, _ := ConstructOpenAPIReq(datasetID, input, op)
	log.Println("constructred openapi request: ", openapiReq)
	openapiResp, err := policyManager.GetPoliciesDecisions(openapiReq, creds)
	log.Println("openapi response received from policy manager: ", openapiResp)
	var actions []*openapiclientmodels.ResultItem
	result := openapiResp.GetResult()
	for i := 0; i < len(result); i++ {
		actions = append(actions, &result[i])
	}
	return actions, err
}
