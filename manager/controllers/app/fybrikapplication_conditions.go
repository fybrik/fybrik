// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"strings"

	corev1 "k8s.io/api/core/v1"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/logging"
)

// Condition indices are static. Conditions always present in the status.
const (
	// ReadyCondition means that access to a dataset is granted
	ReadyConditionIndex int64 = 0
	// DenyCondition means that access to a dataset is denied
	DenyConditionIndex int64 = 1
	// ErrorCondition means that an error was encountered during blueprint construction
	ErrorConditionIndex int64 = 2
	numConditions       int   = 3
)

// Helper functions to manage conditions

func initStatus(application *fapp.FybrikApplication) {
	application.Status.ErrorMessage = ""
	application.Status.AssetStates = make(map[string]fapp.AssetState)
	if len(application.Spec.Data) == 0 {
		application.Status.Ready = true
	} else {
		application.Status.Ready = false
	}
	for _, asset := range application.Spec.Data {
		resetAssetState(application, asset.DataSetID)
	}
}

func resetAssetState(application *fapp.FybrikApplication, assetID string) {
	conditions := make([]fapp.Condition, numConditions)
	conditions[ErrorConditionIndex] = fapp.Condition{Type: fapp.ErrorCondition, Status: corev1.ConditionFalse}
	conditions[DenyConditionIndex] = fapp.Condition{Type: fapp.DenyCondition, Status: corev1.ConditionFalse}
	conditions[ReadyConditionIndex] = fapp.Condition{Type: fapp.ReadyCondition, Status: corev1.ConditionFalse}
	application.Status.AssetStates[assetID] = fapp.AssetState{Conditions: conditions}
}

func setErrorCondition(appContext ApplicationContext, assetID, msg string) {
	appContext.Application.Status.AssetStates[assetID].Conditions[ErrorConditionIndex] = fapp.Condition{
		Type:    fapp.ErrorCondition,
		Status:  corev1.ConditionTrue,
		Message: msg}
	appContext.Log.Error().Bool(logging.FORUSER, true).Bool(logging.AUDIT, true).
		Str(logging.DATASETID, assetID).Msgf("Setting error condition: %s", msg)
}

func setDenyCondition(appContext ApplicationContext, assetID, msg string) {
	appContext.Application.Status.AssetStates[assetID].Conditions[DenyConditionIndex] = fapp.Condition{
		Type:    fapp.DenyCondition,
		Status:  corev1.ConditionTrue,
		Message: msg}
	appContext.Log.Error().Bool(logging.FORUSER, true).Bool(logging.AUDIT, true).
		Str(logging.DATASETID, assetID).Msg("Setting deny condition: " + msg)
}

func setReadyCondition(appContext ApplicationContext, assetID string) {
	appContext.Application.Status.AssetStates[assetID].Conditions[ReadyConditionIndex].Status = corev1.ConditionTrue
	appContext.Log.Info().Bool(logging.FORUSER, true).Bool(logging.AUDIT, true).
		Str(logging.DATASETID, assetID).Msg("Setting ready condition")
}

// determine if the application is ready
func isReady(application *fapp.FybrikApplication) bool {
	if len(application.Spec.Data) == 0 {
		return true
	}
	if application.Status.AssetStates == nil {
		return false
	}
	for _, asset := range application.Spec.Data {
		assetState := application.Status.AssetStates[asset.DataSetID]
		if len(assetState.Conditions) == 0 {
			return false
		}
		if assetState.Conditions[DenyConditionIndex].Status == corev1.ConditionFalse &&
			assetState.Conditions[ReadyConditionIndex].Status == corev1.ConditionFalse {
			return false
		}
	}
	return true
}

func getErrorMessages(application *fapp.FybrikApplication) string {
	if application.Status.ErrorMessage != "" {
		return application.Status.ErrorMessage
	}
	var errorMsgs []string
	for _, state := range application.Status.AssetStates {
		if state.Conditions[ErrorConditionIndex].Status == corev1.ConditionTrue {
			errorMsgs = append(errorMsgs, state.Conditions[ErrorConditionIndex].Message)
		}
	}
	return strings.Join(errorMsgs, "\n")
}
