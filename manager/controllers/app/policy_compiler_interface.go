// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"

	"emperror.dev/errors"
	"github.com/gdexlab/go-render/render"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	connectors "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/policymanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/taxonomy/validate"
	"fybrik.io/fybrik/pkg/vault"
)

var PolicyManagerTaxonomy = environment.GetDataDir() + "/taxonomy/policymanager.json#/definitions/GetPolicyDecisionsResponse"

func ConstructOpenAPIReq(datasetID string, resourceMetadata *datacatalog.ResourceMetadata, input *app.FybrikApplication,
	operation *policymanager.RequestAction) *policymanager.GetPolicyDecisionsRequest {
	return &policymanager.GetPolicyDecisionsRequest{
		Context: taxonomy.PolicyManagerRequestContext{Properties: input.Spec.AppInfo.Properties},
		Action:  *operation,
		Resource: policymanager.Resource{
			ID:       taxonomy.AssetID(datasetID),
			Metadata: resourceMetadata.DeepCopy(),
		},
	}
}

func ValidatePolicyDecisionsResponse(response *policymanager.GetPolicyDecisionsResponse, taxonomyFile string) error {
	var allErrs []*field.Error

	// Convert GetAssetRequest Go struct to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return err
	}

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
		response.DecisionID, allErrs)
}

// LookupPolicyDecisions provides a list of governance actions for the given dataset and the given operation
func LookupPolicyDecisions(datasetID string, resourceMetadata *datacatalog.ResourceMetadata,
	policyManager connectors.PolicyManager, appContext ApplicationContext,
	op *policymanager.RequestAction) ([]taxonomy.Action, error) {
	// call external policy manager to get governance instructions for this operation
	openapiReq := ConstructOpenAPIReq(datasetID, resourceMetadata, appContext.Application, op)
	output := render.AsCode(openapiReq)
	appContext.Log.Debug().Str(logging.DATASETID, datasetID).Msgf("request: %s", output)

	var creds string
	if appContext.Application.Spec.SecretRef != "" {
		if !environment.IsVaultEnabled() {
			appContext.Log.Error().Str("SecretRef", appContext.Application.Spec.SecretRef).Msg("SecretRef defined [%s], but vault is disabled")
		} else {
			creds = vault.PathForReadingKubeSecret(appContext.Application.Namespace, appContext.Application.Spec.SecretRef)
		}
	}

	openapiResp, err := policyManager.GetPoliciesDecisions(openapiReq, creds)
	var actions []taxonomy.Action
	if err != nil {
		return actions, err
	}

	err = ValidatePolicyDecisionsResponse(openapiResp, PolicyManagerTaxonomy)
	if err != nil {
		appContext.Log.Error().Err(err).Str(logging.DATASETID, datasetID).Msg("error while validating policy manager response")
		return actions, errors.New("Validation error: " + err.Error())
	}

	output = render.AsCode(openapiResp)
	appContext.Log.Info().Str(logging.DATASETID, datasetID).Msgf("response from policy manager: %s", output)

	result := openapiResp.Result
	for i := 0; i < len(result); i++ {
		if utils.IsDenied(result[i].Action.Name) {
			var message string
			switch openapiReq.Action.ActionType {
			case taxonomy.ReadFlow:
				message = app.ReadAccessDenied
			case taxonomy.WriteFlow:
				message = app.WriteNotAllowed
			}
			return actions, errors.New(message)
		}
		actions = append(actions, result[i].Action)
	}
	return actions, nil
}
