// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminrules

import "fybrik.io/fybrik/pkg/model/taxonomy"

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
type StringList []string

type RangeType struct {
	Min int `json:"min,omitempty"`
	Max int `json:"max,omitempty"`
}

type Restriction struct {
	Property string     `json:"property"`
	Values   StringList `json:"values,omitempty"`
	Range    *RangeType `json:"range,omitempty"`
}

// DecisionPolicy is a justification for a policy that consists of a unique id, id of a policy set and a human readable desciption
type DecisionPolicy struct {
	ID          string `json:"ID"`
	PolicySetID string `json:"policySetID,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

// Deployment restrictions on modules, clusters and additional resources that will be added in the future
type Restrictions struct {
	Clusters        []Restriction `json:"clusters,omitempty"`
	Modules         []Restriction `json:"modules,omitempty"`
	StorageAccounts []Restriction `json:"storageaccounts,omitempty"`
}

// Decision is a result of evaluating a configuration policy which satisfies the specified predicates
type Decision struct {
	// a decision regarding deployment: True = require, False = forbid, Unknown = allow
	Deploy DeploymentStatus `json:"deploy,omitempty"`
	// Deployment restrictions on modules, clusters and additional resources
	DeploymentRestrictions Restrictions `json:"restrictions,omitempty"`
	// Descriptions of policies that have been used for evaluation
	Policy DecisionPolicy `json:"policy,omitempty"`
}

type DecisionPerCapability struct {
	Capability taxonomy.Capability `json:"capability"`
	Decision   Decision            `json:"decision"`
}

// A list of decisions,
// e.g. [{"capability": "read", "decision": {"deploy": "True"}}, {"capability": "write", "decision": {"deploy": "False"}}]
type RuleDecisionList []DecisionPerCapability

// Result of query evaluation
type EvaluationOutputStructure struct {
	Config RuleDecisionList `json:"config"`
}
