// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// ConditionType represents a condition type of m4dapplication and blueprint resources
type ConditionType string

const (
	// ReadyCondition means that data endpoints are available for usage
	ReadyCondition ConditionType = "Ready"

	// ErrorCondition means that an error was encountered, e.g. failure to establish communication with an external service
	// TODO: add a number of retrials, not activate failure immediately
	ErrorCondition ConditionType = "Error"

	// FailureCondition represents a critical failure, e.g. denial of data access or failure of an allocated resource
	FailureCondition ConditionType = "Failure"

	// TerminatingCondition means that deletion is in progress
	TerminatingCondition ConditionType = "Terminating"
)

// Condition describes the state of a M4DApplication at a certain point.
type Condition struct {
	// Type of the condition
	Type ConditionType `json:"type"`
	// Status of the condition: true,false,unknown
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a short explanation of the reason for the current condition
	// +optional
	Reason string `json:"reason,omitempty"`
	// Message contains the details of the current condition
	// +optional
	Message string `json:"message,omitempty"`
}
