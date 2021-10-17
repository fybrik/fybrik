// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// Decision is a result of evaluating a configuration policy which satisfies the specified predicates
type Decision struct {
	// a decision regarding deployment: True = require, False = forbid, Unknown = allow
	Deploy corev1.ConditionStatus
	// Deployment clusters
	Clusters []string
	// Deployment restrictions, e.g. type = plugin
	Restrictions map[string]string
	// Descriptions of policies that have been used for evaluation
	Jusifications []string
}

// EvaluatorOutput is an output of ConfigurationPoliciesEvaluator.
// Used by manager to decide which modules are deployed and in which cluster.
type EvaluatorOutput struct {
	// Valid is true when there is no conflict between the decisions, and false otherwise
	Valid bool
	// Dataset identifier
	DatasetID string
	// Decisions per capability (after being merged)
	ConfigDecisions map[api.CapabilityType]Decision
}
