// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

// Document represents a taxonomy schema document.
type Document struct {
	Schema `json:",inline"`

	SchemaVersion string                `json:"$schema,omitempty"`
	Definitions   map[string]*SchemaRef `json:"definitions,omitempty"`
}

// Deref dereferences a $ref to the schema model that it points to
func (d *Document) Deref(in *SchemaRef) *SchemaRef {
	// TODO: support references across documents?
	if in != nil && in.Ref != "" {
		refName := in.RefName()
		return d.Definitions[refName]
	}
	return in
}

// ToJSONSchemaProps creates a JSONSchemaProps from the document
func (document *Document) ToJSONSchemaProps() *apiextensions.JSONSchemaProps {
	if document == nil {
		return nil
	}
	props := document.Schema.ToJSONSchemaProps(nil)
	props.Schema = apiextensions.JSONSchemaURL(document.SchemaVersion)
	props.Definitions = make(apiextensions.JSONSchemaDefinitions)
	for key, value := range document.Definitions {
		props.Definitions[key] = *value.ToJSONSchemaProps(nil)
	}
	return props
}

// ToFlatJSONSchemaProps creates a JSONSchemaProps from the definitions section in the document
func (document *Document) ToFlatJSONSchemaProps() *apiextensions.JSONSchemaProps {
	if document == nil {
		return nil
	}

	props := document.Schema.ToJSONSchemaProps(document)
	props.Type = "object"
	props.Schema = apiextensions.JSONSchemaURL(document.SchemaVersion)
	props.Properties = make(map[string]apiextensions.JSONSchemaProps, len(document.Definitions))
	for key, value := range document.Definitions {
		props.Properties[key] = *value.ToJSONSchemaProps(document)
	}
	return props
}
