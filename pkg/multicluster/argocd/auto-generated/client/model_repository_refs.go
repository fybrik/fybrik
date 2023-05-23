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

// RepositoryRefs struct for RepositoryRefs
type RepositoryRefs struct {
	Branches *[]string `json:"branches,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
}

// NewRepositoryRefs instantiates a new RepositoryRefs object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRepositoryRefs() *RepositoryRefs {
	this := RepositoryRefs{}
	return &this
}

// NewRepositoryRefsWithDefaults instantiates a new RepositoryRefs object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRepositoryRefsWithDefaults() *RepositoryRefs {
	this := RepositoryRefs{}
	return &this
}

// GetBranches returns the Branches field value if set, zero value otherwise.
func (o *RepositoryRefs) GetBranches() []string {
	if o == nil || o.Branches == nil {
		var ret []string
		return ret
	}
	return *o.Branches
}

// GetBranchesOk returns a tuple with the Branches field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RepositoryRefs) GetBranchesOk() (*[]string, bool) {
	if o == nil || o.Branches == nil {
		return nil, false
	}
	return o.Branches, true
}

// HasBranches returns a boolean if a field has been set.
func (o *RepositoryRefs) HasBranches() bool {
	if o != nil && o.Branches != nil {
		return true
	}

	return false
}

// SetBranches gets a reference to the given []string and assigns it to the Branches field.
func (o *RepositoryRefs) SetBranches(v []string) {
	o.Branches = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *RepositoryRefs) GetTags() []string {
	if o == nil || o.Tags == nil {
		var ret []string
		return ret
	}
	return *o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RepositoryRefs) GetTagsOk() (*[]string, bool) {
	if o == nil || o.Tags == nil {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *RepositoryRefs) HasTags() bool {
	if o != nil && o.Tags != nil {
		return true
	}

	return false
}

// SetTags gets a reference to the given []string and assigns it to the Tags field.
func (o *RepositoryRefs) SetTags(v []string) {
	o.Tags = &v
}

func (o RepositoryRefs) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Branches != nil {
		toSerialize["branches"] = o.Branches
	}
	if o.Tags != nil {
		toSerialize["tags"] = o.Tags
	}
	return json.Marshal(toSerialize)
}

type NullableRepositoryRefs struct {
	value *RepositoryRefs
	isSet bool
}

func (v NullableRepositoryRefs) Get() *RepositoryRefs {
	return v.value
}

func (v *NullableRepositoryRefs) Set(val *RepositoryRefs) {
	v.value = val
	v.isSet = true
}

func (v NullableRepositoryRefs) IsSet() bool {
	return v.isSet
}

func (v *NullableRepositoryRefs) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRepositoryRefs(val *RepositoryRefs) *NullableRepositoryRefs {
	return &NullableRepositoryRefs{value: val, isSet: true}
}

func (v NullableRepositoryRefs) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRepositoryRefs) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}