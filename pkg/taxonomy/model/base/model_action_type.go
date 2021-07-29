/*
 * Policy Manager Service
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 1.0.0
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package base

import (
	"encoding/json"
	"fmt"
)

// ActionType the model 'ActionType'
type ActionType string

// List of ActionType
const (
	READ ActionType = "read"
	WRITE ActionType = "write"
	DELETE ActionType = "delete"
)

var allowedActionTypeEnumValues = []ActionType{
	"read",
	"write",
	"delete",
}

func (v *ActionType) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := ActionType(value)
	for _, existing := range allowedActionTypeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid ActionType", value)
}

// NewActionTypeFromValue returns a pointer to a valid ActionType
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewActionTypeFromValue(v string) (*ActionType, error) {
	ev := ActionType(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for ActionType: valid values are %v", v, allowedActionTypeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v ActionType) IsValid() bool {
	for _, existing := range allowedActionTypeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to ActionType value
func (v ActionType) Ptr() *ActionType {
	return &v
}

type NullableActionType struct {
	value *ActionType
	isSet bool
}

func (v NullableActionType) Get() *ActionType {
	return v.value
}

func (v *NullableActionType) Set(val *ActionType) {
	v.value = val
	v.isSet = true
}

func (v NullableActionType) IsSet() bool {
	return v.isSet
}

func (v *NullableActionType) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableActionType(val *ActionType) *NullableActionType {
	return &NullableActionType{value: val, isSet: true}
}

func (v NullableActionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableActionType) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

