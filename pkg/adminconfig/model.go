// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"reflect"
	"strconv"
	"strings"

	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/utils"
)

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

type Restriction struct {
	Property string              `json:"property"`
	Values   StringList          `json:"values,omitempty"`
	Range    *taxonomy.RangeType `json:"range,omitempty"`
}

// DecisionPolicy is a justification for a policy that consists of a unique id, id of a policy set and a human readable description
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

// +kubebuilder:validation:Enum=min;max
type OptimizationDirective string

// List of directives
const (
	Minimize OptimizationDirective = "min"
	Maximize OptimizationDirective = "max"
)

type AttributeOptimization struct {
	// Attribute name
	// +required
	Attribute string `json:"attribute"`
	// Optimization directive: minimize or maximize
	// +required
	Directive OptimizationDirective `json:"directive"`
	// Weight, a positive number not exceeding 1.0
	// Serialized as a string
	Weight string `json:"weight,omitempty"`
}

// A list of attribute optimizations
type OptimizationStrategy struct {
	Strategy []AttributeOptimization `json:"strategy"`
	Policy   DecisionPolicy          `json:"policy"`
}

// Result of query evaluation
type EvaluationOutputStructure struct {
	Config RuleDecisionList `json:"config"`
	// +optional
	Optimize []OptimizationStrategy `json:"optimize,omitempty"`
}

// Validation of an object with respect to the admin config restriction
//
//nolint:gocyclo
func (restrict Restriction) SatisfiedByResource(attrManager *infrastructure.AttributeManager, spec interface{}, instanceName string) bool {
	details, err := utils.StructToMap(spec)
	if err != nil {
		return false
	}

	var value interface{}
	var found bool
	// infrastructure attribute or a property in the spec?
	value, found = attrManager.GetAttributeValue(restrict.Property, instanceName)
	if !found {
		fields := strings.Split(restrict.Property, ".")
		value, found, err = NestedFieldNoCopy(details, fields...)
	}
	if err != nil || !found {
		return false
	}
	if restrict.Range != nil {
		var numericVal int
		switch value := value.(type) {
		case int64:
			numericVal = int(value)
		case float64:
			numericVal = int(value)
		case int:
			numericVal = value
		case string:
			if numericVal, err = strconv.Atoi(value); err != nil {
				return false
			}
		}
		if restrict.Range.Max > 0 && numericVal > restrict.Range.Max {
			return false
		}
		if restrict.Range.Min > 0 && numericVal < restrict.Range.Min {
			return false
		}
	} else if len(restrict.Values) != 0 {
		if !utils.HasString(value.(string), restrict.Values) {
			return false
		}
	}

	return true
}

func NestedFieldNoCopy(obj map[string]interface{}, fields ...string) (interface{}, bool, error) {
	var val interface{} = obj

	for _, field := range fields {
		if val == nil {
			return nil, false, nil
		}
		if reflect.TypeOf(val).Kind() == reflect.Slice {
			s := reflect.ValueOf(val)
			i, err := strconv.Atoi(field)
			if err != nil {
				return nil, false, nil
			}
			val = s.Index(i).Interface()
			continue
		}
		if m, ok := val.(map[string]interface{}); ok {
			val, ok = m[field]
			if !ok {
				return nil, false, nil
			}
		} else {
			return nil, false, nil
		}
	}
	return val, true, nil
}
