// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Helper functions to manage conditions

func resetConditions(application *app.M4DApplication) {
	application.Status.Ready = false
	application.Status.Conditions = make([]app.Condition, 3)
	application.Status.Conditions[app.ErrorConditionIndex] = app.Condition{Type: app.ErrorCondition, Status: corev1.ConditionFalse}
	application.Status.Conditions[app.DenyConditionIndex] = app.Condition{Type: app.DenyCondition, Status: corev1.ConditionFalse}
	application.Status.Conditions[app.ReadyConditionIndex] = app.Condition{Type: app.ReadyCondition, Status: corev1.ConditionFalse}
}

func setErrorCondition(application *app.M4DApplication, assetID string, msg string) {
	if len(application.Status.Conditions) == 0 {
		resetConditions(application)
	}
	errMsg := "An error was received"
	if assetID != "" {
		errMsg += " for asset " + assetID
	}
	errMsg += " . If the error persists, please contact an operator.\n"
	errMsg += "Error description: " + msg + "\n"
	application.Status.Conditions[app.ErrorConditionIndex].Status = corev1.ConditionTrue
	application.Status.Conditions[app.ErrorConditionIndex].Message += errMsg
}

func setDenyCondition(application *app.M4DApplication, assetID string, msg string) {
	if len(application.Status.Conditions) == 0 {
		resetConditions(application)
	}
	application.Status.Conditions[app.DenyConditionIndex].Status = corev1.ConditionTrue
	application.Status.Conditions[app.DenyConditionIndex].Message = msg
}

func setReadyCondition(application *app.M4DApplication, assetID string) {
	if len(application.Status.Conditions) == 0 {
		resetConditions(application)
	}
	application.Status.Conditions[app.ReadyConditionIndex].Status = corev1.ConditionTrue
	application.Status.Ready = true
}

func hasError(application *app.M4DApplication) bool {
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return false
	}
	return (application.Status.Conditions[app.ErrorConditionIndex].Status == corev1.ConditionTrue ||
		application.Status.Conditions[app.DenyConditionIndex].Status == corev1.ConditionTrue)
}

// Ready or Deny state
func inFinalState(application *app.M4DApplication) bool {
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return false
	}
	return (application.Status.Conditions[app.ReadyConditionIndex].Status == corev1.ConditionTrue ||
		application.Status.Conditions[app.DenyConditionIndex].Status == corev1.ConditionTrue)
}

func getErrorMessages(application *app.M4DApplication) string {
	var errMsg string
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return errMsg
	}
	if application.Status.Conditions[app.ErrorConditionIndex].Status == corev1.ConditionTrue {
		errMsg += application.Status.Conditions[app.ErrorConditionIndex].Message
	}
	if application.Status.Conditions[app.DenyConditionIndex].Status == corev1.ConditionTrue {
		errMsg += application.Status.Conditions[app.DenyConditionIndex].Message
	}
	return errMsg
}
