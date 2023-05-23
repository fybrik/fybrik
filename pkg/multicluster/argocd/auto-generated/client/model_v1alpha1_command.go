/*
Consolidate Services

Description of all APIs

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapiclient

import (
	"encoding/json"
)

// V1alpha1Command struct for V1alpha1Command
type V1alpha1Command struct {
	Args    *[]string `json:"args,omitempty"`
	Command *[]string `json:"command,omitempty"`
}

// NewV1alpha1Command instantiates a new V1alpha1Command object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1alpha1Command() *V1alpha1Command {
	this := V1alpha1Command{}
	return &this
}

// NewV1alpha1CommandWithDefaults instantiates a new V1alpha1Command object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1alpha1CommandWithDefaults() *V1alpha1Command {
	this := V1alpha1Command{}
	return &this
}

// GetArgs returns the Args field value if set, zero value otherwise.
func (o *V1alpha1Command) GetArgs() []string {
	if o == nil || o.Args == nil {
		var ret []string
		return ret
	}
	return *o.Args
}

// GetArgsOk returns a tuple with the Args field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1Command) GetArgsOk() (*[]string, bool) {
	if o == nil || o.Args == nil {
		return nil, false
	}
	return o.Args, true
}

// HasArgs returns a boolean if a field has been set.
func (o *V1alpha1Command) HasArgs() bool {
	if o != nil && o.Args != nil {
		return true
	}

	return false
}

// SetArgs gets a reference to the given []string and assigns it to the Args field.
func (o *V1alpha1Command) SetArgs(v []string) {
	o.Args = &v
}

// GetCommand returns the Command field value if set, zero value otherwise.
func (o *V1alpha1Command) GetCommand() []string {
	if o == nil || o.Command == nil {
		var ret []string
		return ret
	}
	return *o.Command
}

// GetCommandOk returns a tuple with the Command field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1Command) GetCommandOk() (*[]string, bool) {
	if o == nil || o.Command == nil {
		return nil, false
	}
	return o.Command, true
}

// HasCommand returns a boolean if a field has been set.
func (o *V1alpha1Command) HasCommand() bool {
	if o != nil && o.Command != nil {
		return true
	}

	return false
}

// SetCommand gets a reference to the given []string and assigns it to the Command field.
func (o *V1alpha1Command) SetCommand(v []string) {
	o.Command = &v
}

func (o V1alpha1Command) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Args != nil {
		toSerialize["args"] = o.Args
	}
	if o.Command != nil {
		toSerialize["command"] = o.Command
	}
	return json.Marshal(toSerialize)
}

type NullableV1alpha1Command struct {
	value *V1alpha1Command
	isSet bool
}

func (v NullableV1alpha1Command) Get() *V1alpha1Command {
	return v.value
}

func (v *NullableV1alpha1Command) Set(val *V1alpha1Command) {
	v.value = val
	v.isSet = true
}

func (v NullableV1alpha1Command) IsSet() bool {
	return v.isSet
}

func (v *NullableV1alpha1Command) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1alpha1Command(val *V1alpha1Command) *NullableV1alpha1Command {
	return &NullableV1alpha1Command{value: val, isSet: true}
}

func (v NullableV1alpha1Command) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1alpha1Command) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}