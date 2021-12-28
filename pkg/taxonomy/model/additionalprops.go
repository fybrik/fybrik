// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/json"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

// AdditionalPropertiesType represents the additionalProperties field.
// The additionalProperties field can either hold a boolean or a schema.
type AdditionalPropertiesType struct {
	Allowed *bool
	Schema  *SchemaRef
}

// IsAllowed returns true when `additionalProperties: true`
func (a *AdditionalPropertiesType) IsAllowed() bool {
	return a != nil && a.Allowed != nil && *a.Allowed
}

func (a *AdditionalPropertiesType) UnmarshalJSON(data []byte) (err error) {
	result := AdditionalPropertiesType{
		Allowed: nil,
		Schema:  nil,
	}

	if err = json.Unmarshal(data, &result.Allowed); err != nil {
		result.Allowed = nil
		if err = json.Unmarshal(data, &result.Schema); err != nil {
			return err
		}
	}

	*a = result
	return nil
}

func (a AdditionalPropertiesType) MarshalJSON() ([]byte, error) {
	if a.Allowed != nil {
		return json.Marshal(a.Allowed)
	}

	return json.Marshal(a.Schema)
}

func (a *AdditionalPropertiesType) ToJSONSchemaProps(flattenBy *Document) *apiextensions.JSONSchemaPropsOrBool {
	if a == nil {
		return nil
	}

	if a.Schema != nil {
		return &apiextensions.JSONSchemaPropsOrBool{Allows: true, Schema: a.Schema.ToJSONSchemaProps(flattenBy)}
	}

	// additionalProperties: false is forbidden in Kubernetes
	// x-kubernetes-preserve-unknown-fields replaces the functionality of additionalProperties: true
	if a.Allowed != nil {
		// return &apiextensions.JSONSchemaPropsOrBool{Allows: *a.Allowed}
		return nil
	}

	return &apiextensions.JSONSchemaPropsOrBool{Allows: false}
}
