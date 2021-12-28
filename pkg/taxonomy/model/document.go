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
func (d *Document) ToJSONSchemaProps() *apiextensions.JSONSchemaProps {
	if d == nil {
		return nil
	}
	props := d.Schema.ToJSONSchemaProps(nil)
	props.Schema = apiextensions.JSONSchemaURL(d.SchemaVersion)
	props.Definitions = make(apiextensions.JSONSchemaDefinitions)
	for key, value := range d.Definitions {
		props.Definitions[key] = *value.ToJSONSchemaProps(nil)
	}
	return props
}

// ToFlatJSONSchemaProps creates a JSONSchemaProps from the definitions section in the document
func (d *Document) ToFlatJSONSchemaProps() *apiextensions.JSONSchemaProps {
	if d == nil {
		return nil
	}

	props := d.Schema.ToJSONSchemaProps(d)
	props.Type = "object"
	props.Schema = apiextensions.JSONSchemaURL(d.SchemaVersion)
	props.Properties = make(map[string]apiextensions.JSONSchemaProps, len(d.Definitions))
	for key, value := range d.Definitions {
		props.Properties[key] = *value.ToJSONSchemaProps(d)
	}
	return props
}
