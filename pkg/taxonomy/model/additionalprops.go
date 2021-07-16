// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package model

import "encoding/json"

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
