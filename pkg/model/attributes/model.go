// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package infraattributes

import (
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

type RangeType struct {
	Min int `json:"min,omitempty"`
	Max int `json:"max,omitempty"`
}

type InfrastructureElement struct {
	Attribute   taxonomy.Attribute `json:"attribute"`
	Description string             `json:"description,omitempty"`
	Type        AttributeType      `json:"type"`
	Value       string             `json:"value"`
	Units       taxonomy.Units     `json:"units,omitempty"`
	Instance    string             `json:"instance,omitempty"`
	Scale       *RangeType         `json:"scale,omitempty"`
}

type Infrastructure struct {
	Items []InfrastructureElement `json:"infrastructure"`
}
