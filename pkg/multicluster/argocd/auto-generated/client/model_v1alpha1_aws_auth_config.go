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

// V1alpha1AWSAuthConfig struct for V1alpha1AWSAuthConfig
type V1alpha1AWSAuthConfig struct {
	ClusterName *string `json:"clusterName,omitempty"`
	// RoleARN contains optional role ARN. If set then AWS IAM Authenticator assume a role to perform cluster operations instead of the default AWS credential provider chain.
	RoleARN *string `json:"roleARN,omitempty"`
}

// NewV1alpha1AWSAuthConfig instantiates a new V1alpha1AWSAuthConfig object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1alpha1AWSAuthConfig() *V1alpha1AWSAuthConfig {
	this := V1alpha1AWSAuthConfig{}
	return &this
}

// NewV1alpha1AWSAuthConfigWithDefaults instantiates a new V1alpha1AWSAuthConfig object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1alpha1AWSAuthConfigWithDefaults() *V1alpha1AWSAuthConfig {
	this := V1alpha1AWSAuthConfig{}
	return &this
}

// GetClusterName returns the ClusterName field value if set, zero value otherwise.
func (o *V1alpha1AWSAuthConfig) GetClusterName() string {
	if o == nil || o.ClusterName == nil {
		var ret string
		return ret
	}
	return *o.ClusterName
}

// GetClusterNameOk returns a tuple with the ClusterName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1AWSAuthConfig) GetClusterNameOk() (*string, bool) {
	if o == nil || o.ClusterName == nil {
		return nil, false
	}
	return o.ClusterName, true
}

// HasClusterName returns a boolean if a field has been set.
func (o *V1alpha1AWSAuthConfig) HasClusterName() bool {
	if o != nil && o.ClusterName != nil {
		return true
	}

	return false
}

// SetClusterName gets a reference to the given string and assigns it to the ClusterName field.
func (o *V1alpha1AWSAuthConfig) SetClusterName(v string) {
	o.ClusterName = &v
}

// GetRoleARN returns the RoleARN field value if set, zero value otherwise.
func (o *V1alpha1AWSAuthConfig) GetRoleARN() string {
	if o == nil || o.RoleARN == nil {
		var ret string
		return ret
	}
	return *o.RoleARN
}

// GetRoleARNOk returns a tuple with the RoleARN field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1AWSAuthConfig) GetRoleARNOk() (*string, bool) {
	if o == nil || o.RoleARN == nil {
		return nil, false
	}
	return o.RoleARN, true
}

// HasRoleARN returns a boolean if a field has been set.
func (o *V1alpha1AWSAuthConfig) HasRoleARN() bool {
	if o != nil && o.RoleARN != nil {
		return true
	}

	return false
}

// SetRoleARN gets a reference to the given string and assigns it to the RoleARN field.
func (o *V1alpha1AWSAuthConfig) SetRoleARN(v string) {
	o.RoleARN = &v
}

func (o V1alpha1AWSAuthConfig) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.ClusterName != nil {
		toSerialize["clusterName"] = o.ClusterName
	}
	if o.RoleARN != nil {
		toSerialize["roleARN"] = o.RoleARN
	}
	return json.Marshal(toSerialize)
}

type NullableV1alpha1AWSAuthConfig struct {
	value *V1alpha1AWSAuthConfig
	isSet bool
}

func (v NullableV1alpha1AWSAuthConfig) Get() *V1alpha1AWSAuthConfig {
	return v.value
}

func (v *NullableV1alpha1AWSAuthConfig) Set(val *V1alpha1AWSAuthConfig) {
	v.value = val
	v.isSet = true
}

func (v NullableV1alpha1AWSAuthConfig) IsSet() bool {
	return v.isSet
}

func (v *NullableV1alpha1AWSAuthConfig) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1alpha1AWSAuthConfig(val *V1alpha1AWSAuthConfig) *NullableV1alpha1AWSAuthConfig {
	return &NullableV1alpha1AWSAuthConfig{value: val, isSet: true}
}

func (v NullableV1alpha1AWSAuthConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1alpha1AWSAuthConfig) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
