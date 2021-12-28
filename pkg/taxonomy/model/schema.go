// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

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
	Default              *apiextensions.JSON       `json:"default,omitempty"`

	// Composable
	OneOf SchemaRefs `json:"oneOf,omitempty"`
	AnyOf SchemaRefs `json:"anyOf,omitempty"`
	AllOf SchemaRefs `json:"allOf,omitempty"`
	Not   *SchemaRef `json:"not,omitempty"`

	// Object
	Required []string `json:"required,omitempty"`
	MinProps *int64   `json:"minProperties,omitempty"`
	MaxProps *int64   `json:"maxProperties,omitempty"`

	// String
	Format    string               `json:"format,omitempty"`
	Enum      []apiextensions.JSON `json:"enum,omitempty"`
	MinLength *int64               `json:"minLength,omitempty"`
	MaxLength *int64               `json:"maxLength,omitempty"`
	Pattern   string               `json:"pattern,omitempty"`

	// Number
	Min          *float64 `json:"minimum,omitempty"`
	Max          *float64 `json:"maximum,omitempty"`
	MultipleOf   *float64 `json:"multipleOf,omitempty"`
	ExclusiveMin bool     `json:"exclusiveMinimum,omitempty"`
	ExclusiveMax bool     `json:"exclusiveMaximum,omitempty"`

	// Array
	MinItems    *int64 `json:"minItems,omitempty"`
	MaxItems    *int64 `json:"maxItems,omitempty"`
	UniqueItems bool   `json:"uniqueItems,omitempty"`
}

func (schema *Schema) ToJSONSchemaProps(flattenBy *Document) *apiextensions.JSONSchemaProps {
	if schema == nil {
		return nil
	}

	return &apiextensions.JSONSchemaProps{
		Description:          schema.Description,
		Type:                 schema.Type,
		Format:               schema.Format,
		Title:                schema.Title,
		Default:              schema.Default,
		Maximum:              schema.Max,
		ExclusiveMaximum:     schema.ExclusiveMax,
		Minimum:              schema.Min,
		ExclusiveMinimum:     schema.ExclusiveMin,
		MaxLength:            schema.MaxLength,
		MinLength:            schema.MinLength,
		Pattern:              schema.Pattern,
		MaxItems:             schema.MaxItems,
		MinItems:             schema.MinItems,
		UniqueItems:          schema.UniqueItems,
		MultipleOf:           schema.MultipleOf,
		Enum:                 schema.Enum,
		MaxProperties:        schema.MaxProps,
		MinProperties:        schema.MinProps,
		Required:             schema.Required,
		Items:                &apiextensions.JSONSchemaPropsOrArray{Schema: schema.Items.ToJSONSchemaProps(flattenBy)},
		AllOf:                schema.AllOf.ToJSONSchemaProps(flattenBy),
		OneOf:                schema.OneOf.ToJSONSchemaProps(flattenBy),
		AnyOf:                schema.AnyOf.ToJSONSchemaProps(flattenBy),
		Not:                  schema.Not.ToJSONSchemaProps(flattenBy),
		Properties:           schema.Properties.ToJSONSchemaProps(flattenBy),
		AdditionalProperties: schema.AdditionalProperties.ToJSONSchemaProps(flattenBy),
	}
}
