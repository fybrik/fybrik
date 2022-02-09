// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package attributes

import (
	"fybrik.io/fybrik/pkg/model/adminrules"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// +kubebuilder:validation:Enum=numeric;string;bool
type AttributeType string

// List of attribute types
const (
	Numeric AttributeType = "numeric"
	String  AttributeType = "string"
	Bool    AttributeType = "bool"
)

type InfrastructureElement struct {
	Attribute   taxonomy.Attribute    `json:"attribute"`
	Description string                `json:"description,omitempty"`
	Type        AttributeType         `json:"type"`
	Value       string                `json:"value"`
	Units       taxonomy.Units        `json:"units,omitempty"`
	Instance    string                `json:"instance,omitempty"`
	Scale       *adminrules.RangeType `json:"scale,omitempty"`
}

type Infrastructure struct {
	Items []InfrastructureElement `json:"infrastructure"`
}
