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

// GpgkeyGnuPGPublicKeyCreateResponse struct for GpgkeyGnuPGPublicKeyCreateResponse
type GpgkeyGnuPGPublicKeyCreateResponse struct {
	Created *V1alpha1GnuPGPublicKeyList `json:"created,omitempty"`
	Skipped *[]string                   `json:"skipped,omitempty"`
}

// NewGpgkeyGnuPGPublicKeyCreateResponse instantiates a new GpgkeyGnuPGPublicKeyCreateResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGpgkeyGnuPGPublicKeyCreateResponse() *GpgkeyGnuPGPublicKeyCreateResponse {
	this := GpgkeyGnuPGPublicKeyCreateResponse{}
	return &this
}

// NewGpgkeyGnuPGPublicKeyCreateResponseWithDefaults instantiates a new GpgkeyGnuPGPublicKeyCreateResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGpgkeyGnuPGPublicKeyCreateResponseWithDefaults() *GpgkeyGnuPGPublicKeyCreateResponse {
	this := GpgkeyGnuPGPublicKeyCreateResponse{}
	return &this
}

// GetCreated returns the Created field value if set, zero value otherwise.
func (o *GpgkeyGnuPGPublicKeyCreateResponse) GetCreated() V1alpha1GnuPGPublicKeyList {
	if o == nil || o.Created == nil {
		var ret V1alpha1GnuPGPublicKeyList
		return ret
	}
	return *o.Created
}

// GetCreatedOk returns a tuple with the Created field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GpgkeyGnuPGPublicKeyCreateResponse) GetCreatedOk() (*V1alpha1GnuPGPublicKeyList, bool) {
	if o == nil || o.Created == nil {
		return nil, false
	}
	return o.Created, true
}

// HasCreated returns a boolean if a field has been set.
func (o *GpgkeyGnuPGPublicKeyCreateResponse) HasCreated() bool {
	if o != nil && o.Created != nil {
		return true
	}

	return false
}

// SetCreated gets a reference to the given V1alpha1GnuPGPublicKeyList and assigns it to the Created field.
func (o *GpgkeyGnuPGPublicKeyCreateResponse) SetCreated(v V1alpha1GnuPGPublicKeyList) {
	o.Created = &v
}

// GetSkipped returns the Skipped field value if set, zero value otherwise.
func (o *GpgkeyGnuPGPublicKeyCreateResponse) GetSkipped() []string {
	if o == nil || o.Skipped == nil {
		var ret []string
		return ret
	}
	return *o.Skipped
}

// GetSkippedOk returns a tuple with the Skipped field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GpgkeyGnuPGPublicKeyCreateResponse) GetSkippedOk() (*[]string, bool) {
	if o == nil || o.Skipped == nil {
		return nil, false
	}
	return o.Skipped, true
}

// HasSkipped returns a boolean if a field has been set.
func (o *GpgkeyGnuPGPublicKeyCreateResponse) HasSkipped() bool {
	if o != nil && o.Skipped != nil {
		return true
	}

	return false
}

// SetSkipped gets a reference to the given []string and assigns it to the Skipped field.
func (o *GpgkeyGnuPGPublicKeyCreateResponse) SetSkipped(v []string) {
	o.Skipped = &v
}

func (o GpgkeyGnuPGPublicKeyCreateResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Created != nil {
		toSerialize["created"] = o.Created
	}
	if o.Skipped != nil {
		toSerialize["skipped"] = o.Skipped
	}
	return json.Marshal(toSerialize)
}

type NullableGpgkeyGnuPGPublicKeyCreateResponse struct {
	value *GpgkeyGnuPGPublicKeyCreateResponse
	isSet bool
}

func (v NullableGpgkeyGnuPGPublicKeyCreateResponse) Get() *GpgkeyGnuPGPublicKeyCreateResponse {
	return v.value
}

func (v *NullableGpgkeyGnuPGPublicKeyCreateResponse) Set(val *GpgkeyGnuPGPublicKeyCreateResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableGpgkeyGnuPGPublicKeyCreateResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableGpgkeyGnuPGPublicKeyCreateResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGpgkeyGnuPGPublicKeyCreateResponse(val *GpgkeyGnuPGPublicKeyCreateResponse) *NullableGpgkeyGnuPGPublicKeyCreateResponse {
	return &NullableGpgkeyGnuPGPublicKeyCreateResponse{value: val, isSet: true}
}

func (v NullableGpgkeyGnuPGPublicKeyCreateResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGpgkeyGnuPGPublicKeyCreateResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}