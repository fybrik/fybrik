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

// Restriction connects a property to a list of allowed values.
// Semantics of a list is a disjunction of values, i.e. a type can be either plugin or config.
type StringList []string
type RangeType struct {
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
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
	Bandwidth       []Restriction `json:"bandwidth,omitempty"`
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

// A list of decisions, e.g. [{"capability": "read", "decision": {"deploy": "True"}}, {"capability": "write", "decision": {"deploy": "False"}}]
type RuleDecisionList []DecisionPerCapability

// Result of query evaluation
type EvaluationOutputStructure struct {
	Config RuleDecisionList `json:"config"`
}

type MetricsPerRegion map[taxonomy.ProcessingLocation]taxonomy.BandwidthMetric
type BandwidthMatrix struct {
	Properties []taxonomy.Property                              `json:"properties,omitempty"`
	Values     map[taxonomy.ProcessingLocation]MetricsPerRegion `json:"values"`
}

type Storage struct {
	Properties []taxonomy.Property       `json:"properties,omitempty"`
	Values     []taxonomy.StorageAccount `json:"values,omitempty"`
}

// Infrastructure object
type Infrastructure struct {
	Bandwidth       BandwidthMatrix `json:"bandwidth"`
	StorageAccounts Storage         `json:"storageaccounts"`
}

func (in *BandwidthMatrix) DeepCopyInto(out *BandwidthMatrix) {
	*out = *in
	if in.Properties != nil {
		in, out := &in.Properties, &out.Properties
		*out = make([]taxonomy.Property, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make(map[taxonomy.ProcessingLocation]MetricsPerRegion, len(*in))
		for key, val := range *in {
			if val != nil {
				(*out)[key] = val.DeepCopy()
			}
		}
	}
}

func (in *BandwidthMatrix) DeepCopy() *BandwidthMatrix {
	if in == nil {
		return nil
	}
	out := new(BandwidthMatrix)
	in.DeepCopyInto(out)
	return out
}
