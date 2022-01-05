// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"log"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	connectors "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	"fybrik.io/fybrik/pkg/model/policymanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/taxonomy/validate"
	"fybrik.io/fybrik/pkg/vault"
	"github.com/gdexlab/go-render/render"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ConstructOpenAPIReq(datasetID string, input *app.FybrikApplication, operation *policymanager.RequestAction) *policymanager.GetPolicyDecisionsRequest {
	context := make(map[string]interface{}, len(input.Spec.AppInfo))
	for k, v := range input.Spec.AppInfo {
		context[k] = v
	}

	return &policymanager.GetPolicyDecisionsRequest{
		Context: taxonomy.PolicyManagerRequestContext{Properties: serde.Properties{
			Items: context,
		}},
		Action: *operation,
		Resource: policymanager.Resource{
			ID: taxonomy.AssetID(datasetID),
		},
	}
}

func ValidatePolicyDecisionsRequest(request *policymanager.GetPolicyDecisionsRequest, taxonomyFile string) error {
	var allErrs []*field.Error

	// Convert GetAssetRequest Go struct to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}
	log.Println("requestJSON (policy decisions):" + string(requestJSON))

	// Validate Fybrik module against taxonomy
	allErrs, err = validate.TaxonomyCheck(requestJSON, taxonomyFile)
	if err != nil {
		return err
	}

	// Return any error
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.fybrik.io", Kind: "PolicyManager-GetPolicyDecisionsRequest"},
		string(request.Resource.Metadata.Name), allErrs)
}

func ValidatePolicyDecisionsResponse(response *policymanager.GetPolicyDecisionsResponse, taxonomyFile string) error {
	var allErrs []*field.Error

	// Convert GetAssetRequest Go struct to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return err
	}
	log.Println("responseJSON (policy decisions):" + string(responseJSON))

	// Validate Fybrik module against taxonomy
	allErrs, err = validate.TaxonomyCheck(responseJSON, taxonomyFile)
	if err != nil {
		return err
	}

	// Return any error
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.fybrik.io", Kind: "PolicyManager-GetPolicyDecisionsResponse"},
		string(response.DecisionID), allErrs)
}

// LookupPolicyDecisions provides a list of governance actions for the given dataset and the given operation
func LookupPolicyDecisions(datasetID string, policyManager connectors.PolicyManager, input *app.FybrikApplication, op *policymanager.RequestAction) ([]taxonomy.Action, error) {
	// call external policy manager to get governance instructions for this operation
	openapiReq := ConstructOpenAPIReq(datasetID, input, op)
	output := render.AsCode(openapiReq)
	log.Println("constructed openapi request: ", output)

	var creds string
	if input.Spec.SecretRef != "" {
		creds = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}
	taxonomyFile := "/tmp/taxonomy/policymanager.json#/definitions/GetPolicyDecisionsRequest"
	err := ValidatePolicyDecisionsRequest(openapiReq, taxonomyFile)
	if err != nil {
		return nil, err
	}

	openapiResp, err := policyManager.GetPoliciesDecisions(openapiReq, creds)
	var actions []taxonomy.Action
	if err != nil {
		return actions, err
	}

	taxonomyFile = "/tmp/taxonomy/policymanager.json#/definitions/GetPolicyDecisionsResponse"
	err = ValidatePolicyDecisionsResponse(openapiResp, taxonomyFile)
	if err != nil {
		return actions, errors.New("Validation error")
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
