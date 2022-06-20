// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

type Units string

// +kubebuilder:validation:Enum=numeric;string;bool
type AttributeType string

// List of attribute types
const (
	Numeric AttributeType = "numeric"
	String  AttributeType = "string"
	Bool    AttributeType = "bool"
)

// +kubebuilder:validation:Enum=fybrikmodule;fybrikstorageaccount;cluster
type InstanceType string

// List of instance types
const (
	Module         InstanceType = "fybrikmodule"
	Cluster        InstanceType = "cluster"
	StorageAccount InstanceType = "fybrikstorageaccount"
)

type RangeType struct {
	Min int `json:"min,omitempty"`
	Max int `json:"max,omitempty"`
}

type InfrastructureMetrics struct {
	Name string `json:"name"`
	// Attribute type, e.g. numeric or string
	Type AttributeType `json:"type"`
	// Measurement units
	Units Units `json:"units,omitempty"`
	// A scale of values (minimum and maximum) when applicable
	Scale *RangeType `json:"scale,omitempty"`
}

type InfrastructureElement struct {
	// Attribute name defined in the taxonomy
	Name string `json:"attribute"`
	// Description
	Description string `json:"description,omitempty"`
	// Name of the metric specified in the metrics section
	MetricName string `json:"metricName,omitempty"`
	// Attribute value
	Value string `json:"value"`
	// A resource defined by the attribute ("fybrikstorageaccount","fybrikmodule","cluster")
	Object InstanceType `json:"object,omitempty"`
	// A reference to the resource instance, e.g. storage account name
	Instance string `json:"instance,omitempty"`
	// A list of arguments defining a specific metric, e.g. regions for a bandwidth
	Arguments []string `json:"arguments,omitempty"`
}
