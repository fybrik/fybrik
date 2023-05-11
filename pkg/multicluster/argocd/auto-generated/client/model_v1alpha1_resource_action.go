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

// V1alpha1ResourceAction struct for V1alpha1ResourceAction
type V1alpha1ResourceAction struct {
	Disabled *bool                          `json:"disabled,omitempty"`
	Name     *string                        `json:"name,omitempty"`
	Params   *[]V1alpha1ResourceActionParam `json:"params,omitempty"`
}

// NewV1alpha1ResourceAction instantiates a new V1alpha1ResourceAction object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1alpha1ResourceAction() *V1alpha1ResourceAction {
	this := V1alpha1ResourceAction{}
	return &this
}

// NewV1alpha1ResourceActionWithDefaults instantiates a new V1alpha1ResourceAction object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1alpha1ResourceActionWithDefaults() *V1alpha1ResourceAction {
	this := V1alpha1ResourceAction{}
	return &this
}

// GetDisabled returns the Disabled field value if set, zero value otherwise.
func (o *V1alpha1ResourceAction) GetDisabled() bool {
	if o == nil || o.Disabled == nil {
		var ret bool
		return ret
	}
	return *o.Disabled
}

// GetDisabledOk returns a tuple with the Disabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ResourceAction) GetDisabledOk() (*bool, bool) {
	if o == nil || o.Disabled == nil {
		return nil, false
	}
	return o.Disabled, true
}

// HasDisabled returns a boolean if a field has been set.
func (o *V1alpha1ResourceAction) HasDisabled() bool {
	if o != nil && o.Disabled != nil {
		return true
	}

	return false
}

// SetDisabled gets a reference to the given bool and assigns it to the Disabled field.
func (o *V1alpha1ResourceAction) SetDisabled(v bool) {
	o.Disabled = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *V1alpha1ResourceAction) GetName() string {
	if o == nil || o.Name == nil {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ResourceAction) GetNameOk() (*string, bool) {
	if o == nil || o.Name == nil {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *V1alpha1ResourceAction) HasName() bool {
	if o != nil && o.Name != nil {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *V1alpha1ResourceAction) SetName(v string) {
	o.Name = &v
}

// GetParams returns the Params field value if set, zero value otherwise.
func (o *V1alpha1ResourceAction) GetParams() []V1alpha1ResourceActionParam {
	if o == nil || o.Params == nil {
		var ret []V1alpha1ResourceActionParam
		return ret
	}
	return *o.Params
}

// GetParamsOk returns a tuple with the Params field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ResourceAction) GetParamsOk() (*[]V1alpha1ResourceActionParam, bool) {
	if o == nil || o.Params == nil {
		return nil, false
	}
	return o.Params, true
}

// HasParams returns a boolean if a field has been set.
func (o *V1alpha1ResourceAction) HasParams() bool {
	if o != nil && o.Params != nil {
		return true
	}

	return false
}

// SetParams gets a reference to the given []V1alpha1ResourceActionParam and assigns it to the Params field.
func (o *V1alpha1ResourceAction) SetParams(v []V1alpha1ResourceActionParam) {
	o.Params = &v
}

func (o V1alpha1ResourceAction) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Disabled != nil {
		toSerialize["disabled"] = o.Disabled
	}
	if o.Name != nil {
		toSerialize["name"] = o.Name
	}
	if o.Params != nil {
		toSerialize["params"] = o.Params
	}
	return json.Marshal(toSerialize)
}

type NullableV1alpha1ResourceAction struct {
	value *V1alpha1ResourceAction
	isSet bool
}

func (v NullableV1alpha1ResourceAction) Get() *V1alpha1ResourceAction {
	return v.value
}

func (v *NullableV1alpha1ResourceAction) Set(val *V1alpha1ResourceAction) {
	v.value = val
	v.isSet = true
}

func (v NullableV1alpha1ResourceAction) IsSet() bool {
	return v.isSet
}

func (v *NullableV1alpha1ResourceAction) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1alpha1ResourceAction(val *V1alpha1ResourceAction) *NullableV1alpha1ResourceAction {
	return &NullableV1alpha1ResourceAction{value: val, isSet: true}
}

func (v NullableV1alpha1ResourceAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1alpha1ResourceAction) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
