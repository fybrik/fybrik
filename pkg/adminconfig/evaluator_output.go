// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// DecisionPolicy is a justification for a policy that consists of a unique id and a human readable desciption
type DecisionPolicy struct {
	ID          string `json:"ID,omitempty"`
	Description string `json:"description,omitempty"`
}

// Deployment restrictions on modules, clusters and additional resources
type Restrictions struct {
	// Restrictions on clusters used for deployment
	Clusters []string `json:"clusters,omitempty"`
	// Restrictions on modules of the type “key”: “value” when the key is a module property (e.g. scope, type) and the value is an allowed value (e.g. asset, plugin)
	ModuleRestrictions map[string]string `json:"modules,omitempty"`
}

// Decision is a result of evaluating a configuration policy which satisfies the specified predicates
type Decision struct {
	// a decision regarding deployment: True = require, False = forbid, Unknown = allow
	Deploy corev1.ConditionStatus `json:"deploy,omitempty"`
	// Deployment restrictions on modules, clusters and additional resources
	DeploymentRestrictions Restrictions `json:"restrictions,omitempty"`
	// Descriptions of policies that have been used for evaluation
	Policy DecisionPolicy `json:"policy,omitempty"`
}

type DecisionPerCapabilityMap map[api.CapabilityType]Decision

// EvaluatorOutput is an output of ConfigurationPoliciesEvaluator.
// Used by manager to decide which modules are deployed and in which cluster.
type EvaluatorOutput struct {
	// Valid is true when there is no conflict between the decisions, and false otherwise
	Valid bool
	// Dataset identifier
	DatasetID string
	// Decisions per capability (after being merged)
	ConfigDecisions DecisionPerCapabilityMap
}
