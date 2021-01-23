// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Helper functions to manage conditions

func resetConditions(application *app.M4DApplication) {
	application.Status.Conditions = make([]app.Condition, 2)
	application.Status.Conditions[app.ErrorConditionIndex] = app.Condition{Type: app.ErrorCondition, Status: corev1.ConditionFalse}
	application.Status.Conditions[app.FailureConditionIndex] = app.Condition{Type: app.FailureCondition, Status: corev1.ConditionFalse}
}

func setCondition(application *app.M4DApplication, assetID string, msg string, fatalError bool) {
	if len(application.Status.Conditions) == 0 {
		resetConditions(application)
	}
	errMsg := "An error was received"
	if assetID != "" {
		errMsg += " for asset " + assetID + " . "
	}
	if !fatalError {
		errMsg += "If the error persists, please contact an operator.\n"
	}
	errMsg += "Error description: " + msg + "\n"
	var ind int64
	if fatalError {
		ind = app.FailureConditionIndex
	} else {
		ind = app.ErrorConditionIndex
	}
	application.Status.Conditions[ind].Status = corev1.ConditionTrue
	application.Status.Conditions[ind].Message += errMsg
}

func hasError(application *app.M4DApplication) bool {
	// check if the conditions have been initialized
	if len(application.Status.Conditions) == 0 {
		return false
	}
	return (application.Status.Conditions[app.ErrorConditionIndex].Status == corev1.ConditionTrue ||
		application.Status.Conditions[app.FailureConditionIndex].Status == corev1.ConditionTrue)
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
	if application.Status.Conditions[app.FailureConditionIndex].Status == corev1.ConditionTrue {
		errMsg += application.Status.Conditions[app.FailureConditionIndex].Message
	}
	return errMsg
}
