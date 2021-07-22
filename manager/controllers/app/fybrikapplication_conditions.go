// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"strings"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Helper functions to manage conditions

func resetConditions(application *app.FybrikApplication) {
	application.Status.Ready = false
	application.Status.Conditions = make([]app.Condition, 3)
	application.Status.Conditions[app.ErrorConditionIndex] = app.Condition{Type: app.ErrorCondition, Status: corev1.ConditionFalse}
	application.Status.Conditions[app.DenyConditionIndex] = app.Condition{Type: app.DenyCondition, Status: corev1.ConditionFalse}
	application.Status.Conditions[app.ReadyConditionIndex] = app.Condition{Type: app.ReadyCondition, Status: corev1.ConditionFalse}
}

func setErrorCondition(application *app.FybrikApplication, assetID string, msg string) {
	errMsg := "An error was received"
	if assetID != "" {
		errMsg += " for asset " + assetID
	}
	errMsg += " . If the error persists, please contact an operator."
	errMsg += "Error description: " + msg
	application.Status.Conditions[app.ErrorConditionIndex].Status = corev1.ConditionTrue
	application.Status.Conditions[app.ErrorConditionIndex].Message = errMsg
}

func setDenyCondition(application *app.FybrikApplication, assetID string, msg string) {
	application.Status.Conditions[app.DenyConditionIndex].Status = corev1.ConditionTrue
	application.Status.Conditions[app.DenyConditionIndex].Message = msg
}

func setReadyCondition(application *app.FybrikApplication, assetID string) {
	application.Status.Conditions[app.ReadyConditionIndex].Status = corev1.ConditionTrue
	application.Status.Ready = true
}

func errorOrDeny(application *app.FybrikApplication) bool {
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return false
	}
	return (application.Status.Conditions[app.ErrorConditionIndex].Status == corev1.ConditionTrue ||
		application.Status.Conditions[app.DenyConditionIndex].Status == corev1.ConditionTrue)
}

// Ready or Deny state
func inFinalState(application *app.FybrikApplication) bool {
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return false
	}
	return (application.Status.Conditions[app.ReadyConditionIndex].Status == corev1.ConditionTrue ||
		application.Status.Conditions[app.DenyConditionIndex].Status == corev1.ConditionTrue)
}

func getErrorMessages(application *app.FybrikApplication) string {
	var errorMsgs []string
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return ""
	}
	if application.Status.Conditions[app.ErrorConditionIndex].Status == corev1.ConditionTrue {
		errorMsgs = append(errorMsgs, application.Status.Conditions[app.ErrorConditionIndex].Message)
	}
	if application.Status.Conditions[app.DenyConditionIndex].Status == corev1.ConditionTrue {
		errorMsgs = append(errorMsgs, application.Status.Conditions[app.DenyConditionIndex].Message)
	}
	return strings.Join(errorMsgs, "\n")
}
