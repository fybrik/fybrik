// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policymanager

import "fybrik.io/fybrik/pkg/model/taxonomy"

// +kubebuilder:validation:Enum=read;write;delete
type RequestActionType string

// List of operationType
const (
	READ   RequestActionType = "read"
	WRITE  RequestActionType = "write"
	DELETE RequestActionType = "delete"
)

type RequestAction struct {
	ActionType         RequestActionType           `json:"action_type"`
	ProcessingLocation taxonomy.ProcessingLocation `json:"processing_location,omitempty"`
	Destination        string                      `json:"destination,omitempty"`
}

type ResultItem struct {
	// The policy on which the decision was based
	Policy string          `json:"policy"`
	Action taxonomy.Action `json:"action"`
}
