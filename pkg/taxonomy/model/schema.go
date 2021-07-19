// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package model

// Schema is specified by OpenAPI/Swagger 3.0 standard
// The following fields are removed because they are not defined in JSON Schema DRAFT4 standard:
// Example, ExternalDocs, Nullable, ReadOnly, WriteOnly, AllowEmptyValue, XML, Deprecated, Discriminator.
type Schema struct {
	// Structure
	Type                 string                    `json:"type,omitempty"`
	Title                string                    `json:"title,omitempty"`
	Description          string                    `json:"description,omitempty"`
	Properties           Schemas                   `json:"properties,omitempty"`
	AdditionalProperties *AdditionalPropertiesType `json:"additionalProperties,omitempty"`
	Items                *SchemaRef                `json:"items,omitempty"`
	Default              interface{}               `json:"default,omitempty"`

	// Composable
	OneOf SchemaRefs `json:"oneOf,omitempty"`
	AnyOf SchemaRefs `json:"anyOf,omitempty"`
	AllOf SchemaRefs `json:"allOf,omitempty"`
	Not   *SchemaRef `json:"not,omitempty"`

	// Object
	Required []string `json:"required,omitempty"`
	MinProps uint64   `json:"minProperties,omitempty"`
	MaxProps *uint64  `json:"maxProperties,omitempty"`

	// String
	Format    string        `json:"format,omitempty"`
	Enum      []interface{} `json:"enum,omitempty"`
	MinLength uint64        `json:"minLength,omitempty"`
	MaxLength *uint64       `json:"maxLength,omitempty"`
	Pattern   string        `json:"pattern,omitempty"`

	// Number
	Min          *float64 `json:"minimum,omitempty"`
	Max          *float64 `json:"maximum,omitempty"`
	MultipleOf   *float64 `json:"multipleOf,omitempty"`
	ExclusiveMin bool     `json:"exclusiveMinimum,omitempty"`
	ExclusiveMax bool     `json:"exclusiveMaximum,omitempty"`

	// Array
	MinItems    uint64  `json:"minItems,omitempty"`
	MaxItems    *uint64 `json:"maxItems,omitempty"`
	UniqueItems bool    `json:"uniqueItems,omitempty"`
}
