// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v12

import (
	corev1 "k8s.io/api/core/v1"
)

// ConditionType represents a condition type
type ConditionType string

// Constants defining condition types
const (
	ErrorCondition ConditionType = "Error"
	DenyCondition  ConditionType = "Deny"
	ReadyCondition ConditionType = "Ready"
	ValidCondition ConditionType = "Valid"
)

// Condition describes the state of a FybrikApplication at a certain point.
type Condition struct {
	// Type of the condition
	Type ConditionType `json:"type"`
	// Status of the condition, one of (`True`, `False`, `Unknown`).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=True;False;Unknown
	// +kubebuilder:default:=Unknown
	Status corev1.ConditionStatus `json:"status"`
	// Message contains the details of the current condition
	// +optional
	Message string `json:"message,omitempty"`
	// ObservedGeneration is the version of the resource for which the condition has been evaluated
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}
