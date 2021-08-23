// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"strings"

	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Helper functions to manage conditions

func initStatus(application *api.FybrikApplication) {
	application.Status.ErrorMessage = ""
	application.Status.AssetStates = make(map[string]api.AssetState)
	if len(application.Spec.Data) == 0 {
		application.Status.Ready = true
	} else {
		application.Status.Ready = false
	}
	for _, asset := range application.Spec.Data {
		resetAssetState(application, asset.DataSetID)
	}
}

func resetAssetState(application *api.FybrikApplication, assetID string) {
	conditions := make([]api.Condition, 3)
	conditions[api.ErrorConditionIndex] = api.Condition{Type: api.ErrorCondition, Status: corev1.ConditionFalse}
	conditions[api.DenyConditionIndex] = api.Condition{Type: api.DenyCondition, Status: corev1.ConditionFalse}
	conditions[api.ReadyConditionIndex] = api.Condition{Type: api.ReadyCondition, Status: corev1.ConditionFalse}
	application.Status.AssetStates[assetID] = api.AssetState{Conditions: conditions}
}

func setErrorCondition(application *api.FybrikApplication, assetID string, msg string) {
	errMsg := "An error was received for asset " + assetID
	errMsg += " . If the error persists, please contact an operator."
	errMsg += "Error description: " + msg
	application.Status.AssetStates[assetID].Conditions[api.ErrorConditionIndex] = api.Condition{
		Type:    api.ErrorCondition,
		Status:  corev1.ConditionTrue,
		Message: errMsg}
}

func setDenyCondition(application *api.FybrikApplication, assetID string, msg string) {
	application.Status.AssetStates[assetID].Conditions[api.DenyConditionIndex] = api.Condition{
		Type:    api.DenyCondition,
		Status:  corev1.ConditionTrue,
		Message: msg}
}

func setReadyCondition(application *api.FybrikApplication, assetID string) {
	application.Status.AssetStates[assetID].Conditions[api.ReadyConditionIndex].Status = corev1.ConditionTrue
}

// determine if the application is ready
func isReady(application *api.FybrikApplication) bool {
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
		if assetState.Conditions[api.DenyConditionIndex].Status == corev1.ConditionFalse &&
			assetState.Conditions[api.ReadyConditionIndex].Status == corev1.ConditionFalse {
			return false
		}
	}
	return true
}

func getErrorMessages(application *api.FybrikApplication) string {
	if application.Status.ErrorMessage != "" {
		return application.Status.ErrorMessage
	}
	var errorMsgs []string
	for _, state := range application.Status.AssetStates {
		if state.Conditions[api.ErrorConditionIndex].Status == corev1.ConditionTrue {
			errorMsgs = append(errorMsgs, state.Conditions[api.ErrorConditionIndex].Message)
		}
	}
	return strings.Join(errorMsgs, "\n")
}
