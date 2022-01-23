// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

type PolicySetID string

// +kubebuilder:validation:Enum=True;False;Unknown
type DeploymentStatus string

// DeploymentStatus values
const (
	StatusTrue    DeploymentStatus = "True"
	StatusFalse   DeploymentStatus = "False"
	StatusUnknown DeploymentStatus = "Unknown"
)

// Restriction maps a property to a list of allowed values.
// Semantics is a disjunction of values, i.e. a type can be either plugin or config.
type Restriction map[string][]string

// RestrictedEntity is an entity to be restricted, such as clusters, modules, storageaccounts.
type RestrictedEntity string

// Restricted entities
const (
	Modules         RestrictedEntity = "modules"
	Clusters        RestrictedEntity = "clusters"
	StorageAccounts RestrictedEntity = "storageaccounts"
)

// DecisionPolicy is a justification for a policy that consists of a unique id, id of a policy set and a human readable desciption
type DecisionPolicy struct {
	ID          string `json:"ID"`
	PolicySetID string `json:"policySetID,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

// Deployment restrictions on modules, clusters and additional resources that will be added in the future
type Restrictions map[RestrictedEntity]Restriction

// Decision is a result of evaluating a configuration policy which satisfies the specified predicates
type Decision struct {
	// a decision regarding deployment: True = require, False = forbid, Unknown = allow
	Deploy DeploymentStatus `json:"deploy,omitempty"`
	// Deployment restrictions on modules, clusters and additional resources
	DeploymentRestrictions Restrictions `json:"restrictions,omitempty"`
	// Descriptions of policies that have been used for evaluation
	Policy DecisionPolicy `json:"policy,omitempty"`
}

type DecisionPerCapabilityMap map[Capability]Decision

// A list of decisions per capability, e.g. {"read": {"deploy": "True"}, "write": {"deploy": "False"}}
type RuleDecisionList []DecisionPerCapabilityMap

// Result of query evaluation
type EvaluationOutputStructure struct {
	Config RuleDecisionList `json:"config"`
}
