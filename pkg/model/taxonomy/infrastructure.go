// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

type Attribute string
type Units string

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
	Attribute   Attribute     `json:"attribute"`
	Description string        `json:"description,omitempty"`
	Type        AttributeType `json:"type"`
	Value       string        `json:"value"`
	Units       Units         `json:"units,omitempty"`
	Instance    string        `json:"instance,omitempty"`
	Scale       *RangeType    `json:"scale,omitempty"`
}
