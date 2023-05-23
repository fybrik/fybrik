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

// RepositoryHelmAppSpec struct for RepositoryHelmAppSpec
type RepositoryHelmAppSpec struct {
	FileParameters *[]V1alpha1HelmFileParameter `json:"fileParameters,omitempty"`
	Name           *string                      `json:"name,omitempty"`
	Parameters     *[]V1alpha1HelmParameter     `json:"parameters,omitempty"`
	ValueFiles     *[]string                    `json:"valueFiles,omitempty"`
	Values         *string                      `json:"values,omitempty"`
}

// NewRepositoryHelmAppSpec instantiates a new RepositoryHelmAppSpec object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRepositoryHelmAppSpec() *RepositoryHelmAppSpec {
	this := RepositoryHelmAppSpec{}
	return &this
}

// NewRepositoryHelmAppSpecWithDefaults instantiates a new RepositoryHelmAppSpec object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRepositoryHelmAppSpecWithDefaults() *RepositoryHelmAppSpec {
	this := RepositoryHelmAppSpec{}
	return &this
}

// GetFileParameters returns the FileParameters field value if set, zero value otherwise.
func (o *RepositoryHelmAppSpec) GetFileParameters() []V1alpha1HelmFileParameter {
	if o == nil || o.FileParameters == nil {
		var ret []V1alpha1HelmFileParameter
		return ret
	}
	return *o.FileParameters
}

// GetFileParametersOk returns a tuple with the FileParameters field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RepositoryHelmAppSpec) GetFileParametersOk() (*[]V1alpha1HelmFileParameter, bool) {
	if o == nil || o.FileParameters == nil {
		return nil, false
	}
	return o.FileParameters, true
}

// HasFileParameters returns a boolean if a field has been set.
func (o *RepositoryHelmAppSpec) HasFileParameters() bool {
	if o != nil && o.FileParameters != nil {
		return true
	}

	return false
}

// SetFileParameters gets a reference to the given []V1alpha1HelmFileParameter and assigns it to the FileParameters field.
func (o *RepositoryHelmAppSpec) SetFileParameters(v []V1alpha1HelmFileParameter) {
	o.FileParameters = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *RepositoryHelmAppSpec) GetName() string {
	if o == nil || o.Name == nil {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RepositoryHelmAppSpec) GetNameOk() (*string, bool) {
	if o == nil || o.Name == nil {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *RepositoryHelmAppSpec) HasName() bool {
	if o != nil && o.Name != nil {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *RepositoryHelmAppSpec) SetName(v string) {
	o.Name = &v
}

// GetParameters returns the Parameters field value if set, zero value otherwise.
func (o *RepositoryHelmAppSpec) GetParameters() []V1alpha1HelmParameter {
	if o == nil || o.Parameters == nil {
		var ret []V1alpha1HelmParameter
		return ret
	}
	return *o.Parameters
}

// GetParametersOk returns a tuple with the Parameters field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RepositoryHelmAppSpec) GetParametersOk() (*[]V1alpha1HelmParameter, bool) {
	if o == nil || o.Parameters == nil {
		return nil, false
	}
	return o.Parameters, true
}

// HasParameters returns a boolean if a field has been set.
func (o *RepositoryHelmAppSpec) HasParameters() bool {
	if o != nil && o.Parameters != nil {
		return true
	}

	return false
}

// SetParameters gets a reference to the given []V1alpha1HelmParameter and assigns it to the Parameters field.
func (o *RepositoryHelmAppSpec) SetParameters(v []V1alpha1HelmParameter) {
	o.Parameters = &v
}

// GetValueFiles returns the ValueFiles field value if set, zero value otherwise.
func (o *RepositoryHelmAppSpec) GetValueFiles() []string {
	if o == nil || o.ValueFiles == nil {
		var ret []string
		return ret
	}
	return *o.ValueFiles
}

// GetValueFilesOk returns a tuple with the ValueFiles field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RepositoryHelmAppSpec) GetValueFilesOk() (*[]string, bool) {
	if o == nil || o.ValueFiles == nil {
		return nil, false
	}
	return o.ValueFiles, true
}

// HasValueFiles returns a boolean if a field has been set.
func (o *RepositoryHelmAppSpec) HasValueFiles() bool {
	if o != nil && o.ValueFiles != nil {
		return true
	}

	return false
}

// SetValueFiles gets a reference to the given []string and assigns it to the ValueFiles field.
func (o *RepositoryHelmAppSpec) SetValueFiles(v []string) {
	o.ValueFiles = &v
}

// GetValues returns the Values field value if set, zero value otherwise.
func (o *RepositoryHelmAppSpec) GetValues() string {
	if o == nil || o.Values == nil {
		var ret string
		return ret
	}
	return *o.Values
}

// GetValuesOk returns a tuple with the Values field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RepositoryHelmAppSpec) GetValuesOk() (*string, bool) {
	if o == nil || o.Values == nil {
		return nil, false
	}
	return o.Values, true
}

// HasValues returns a boolean if a field has been set.
func (o *RepositoryHelmAppSpec) HasValues() bool {
	if o != nil && o.Values != nil {
		return true
	}

	return false
}

// SetValues gets a reference to the given string and assigns it to the Values field.
func (o *RepositoryHelmAppSpec) SetValues(v string) {
	o.Values = &v
}

func (o RepositoryHelmAppSpec) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.FileParameters != nil {
		toSerialize["fileParameters"] = o.FileParameters
	}
	if o.Name != nil {
		toSerialize["name"] = o.Name
	}
	if o.Parameters != nil {
		toSerialize["parameters"] = o.Parameters
	}
	if o.ValueFiles != nil {
		toSerialize["valueFiles"] = o.ValueFiles
	}
	if o.Values != nil {
		toSerialize["values"] = o.Values
	}
	return json.Marshal(toSerialize)
}

type NullableRepositoryHelmAppSpec struct {
	value *RepositoryHelmAppSpec
	isSet bool
}

func (v NullableRepositoryHelmAppSpec) Get() *RepositoryHelmAppSpec {
	return v.value
}

func (v *NullableRepositoryHelmAppSpec) Set(val *RepositoryHelmAppSpec) {
	v.value = val
	v.isSet = true
}

func (v NullableRepositoryHelmAppSpec) IsSet() bool {
	return v.isSet
}

func (v *NullableRepositoryHelmAppSpec) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRepositoryHelmAppSpec(val *RepositoryHelmAppSpec) *NullableRepositoryHelmAppSpec {
	return &NullableRepositoryHelmAppSpec{value: val, isSet: true}
}

func (v NullableRepositoryHelmAppSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRepositoryHelmAppSpec) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}