// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

// SchemaRef is either a schema or a reference to a schema
type SchemaRef struct {
	Schema
	Ref string `json:"$ref,omitempty"`
}

// SchemaRefs is a list of schemas/references
type SchemaRefs []*SchemaRef

// Schemas is a map of schemas/references
type Schemas map[string]*SchemaRef

// RefName returns the name from a reference.
// For example given a reference `$ref: "#/definitions/MyObject"` the returned value is "MyObject".
func (schemaRef *SchemaRef) RefName() string {
	if schemaRef != nil && schemaRef.Ref != "" {
		return schemaRef.Ref[strings.LastIndex(schemaRef.Ref, "/")+1:]
	}
	return ""
}

func (schemaRef *SchemaRef) ToJSONSchemaProps(flattenBy *Document) *apiextensions.JSONSchemaProps {
	if schemaRef == nil {
		return nil
	}

	if schemaRef.Ref != "" && flattenBy != nil {
		return flattenBy.Deref(schemaRef).ToJSONSchemaProps(flattenBy)
	}

	if schemaRef.Ref != "" {
		return &apiextensions.JSONSchemaProps{
			Ref: &schemaRef.Ref,
		}
	}

	return schemaRef.Schema.ToJSONSchemaProps(flattenBy)
}

func (schemaRefs *SchemaRefs) ToJSONSchemaProps(flattenBy *Document) []apiextensions.JSONSchemaProps {
	result := make([]apiextensions.JSONSchemaProps, len(*schemaRefs))
	for index, schema := range *schemaRefs {
		result[index] = *schema.ToJSONSchemaProps(flattenBy)
	}
	return result
}

func (schemas *Schemas) ToJSONSchemaProps(flattenBy *Document) map[string]apiextensions.JSONSchemaProps {
	result := make(map[string]apiextensions.JSONSchemaProps)
	if schemas != nil {
		for k, v := range *schemas {
			result[k] = *v.ToJSONSchemaProps(flattenBy)
		}
	}
	return result
}
