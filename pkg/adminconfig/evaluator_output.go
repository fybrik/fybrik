// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	corev1 "k8s.io/api/core/v1"
)

// Restriction maps a property to a list of allowed values
// For example, a module restriction can map "type" to ["plugin", "config"], and "scope" to ["workload"]
// These values come from FybrikModule taxonomy.
// Semantics is a disjunction of values, i.e. a type can be either plugin or config
type Restriction map[string][]string

// RestrictedEntity is an entity to be restricted, such as clusters, modules, storageaccounts.
type RestrictedEntity string

// Restricted entities
const (
	Modules         RestrictedEntity = "modules"
	Clusters        RestrictedEntity = "clusters"
	StorageAccounts RestrictedEntity = "storage"
)

// DecisionPolicy is a justification for a policy that consists of a unique id, id of a policy set and a human readable desciption
// TODO(shlomitk1): add effective date, expiration date.
type DecisionPolicy struct {
	ID          string `json:"ID"`
	PolicySetID string `json:"policySetID"`
	Description string `json:"description,omitempty"`
}

// Deployment restrictions on modules, clusters and additional resources that will be added in the future
type Restrictions map[RestrictedEntity]Restriction

// Decision is a result of evaluating a configuration policy which satisfies the specified predicates
type Decision struct {
	// a decision regarding deployment: True = require, False = forbid, Unknown = allow
	Deploy corev1.ConditionStatus `json:"deploy,omitempty"`
	// Deployment restrictions on modules, clusters and additional resources
	DeploymentRestrictions Restrictions `json:"restrictions,omitempty"`
	// Descriptions of policies that have been used for evaluation
	Policy DecisionPolicy `json:"policy,omitempty"`
}

type DecisionPerCapabilityMap map[string]Decision

// EvaluatorOutput is an output of ConfigurationPoliciesEvaluator.
// Used by manager to decide which modules are deployed and in which cluster.
type EvaluatorOutput struct {
	// Valid is true when there is no conflict between the decisions, and false otherwise
	Valid bool
	// Dataset identifier
	DatasetID string
	// Unique fybrikapplication id used for logging
	UUID string
	// Policy set id used in the evaluation
	PolicySetID string
	// Decisions per capability (after being merged)
	ConfigDecisions DecisionPerCapabilityMap
}
