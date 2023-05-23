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

// V1alpha1Operation struct for V1alpha1Operation
type V1alpha1Operation struct {
	Info        *[]V1alpha1Info             `json:"info,omitempty"`
	InitiatedBy *V1alpha1OperationInitiator `json:"initiatedBy,omitempty"`
	Retry       *V1alpha1RetryStrategy      `json:"retry,omitempty"`
	Sync        *V1alpha1SyncOperation      `json:"sync,omitempty"`
}

// NewV1alpha1Operation instantiates a new V1alpha1Operation object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1alpha1Operation() *V1alpha1Operation {
	this := V1alpha1Operation{}
	return &this
}

// NewV1alpha1OperationWithDefaults instantiates a new V1alpha1Operation object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1alpha1OperationWithDefaults() *V1alpha1Operation {
	this := V1alpha1Operation{}
	return &this
}

// GetInfo returns the Info field value if set, zero value otherwise.
func (o *V1alpha1Operation) GetInfo() []V1alpha1Info {
	if o == nil || o.Info == nil {
		var ret []V1alpha1Info
		return ret
	}
	return *o.Info
}

// GetInfoOk returns a tuple with the Info field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1Operation) GetInfoOk() (*[]V1alpha1Info, bool) {
	if o == nil || o.Info == nil {
		return nil, false
	}
	return o.Info, true
}

// HasInfo returns a boolean if a field has been set.
func (o *V1alpha1Operation) HasInfo() bool {
	if o != nil && o.Info != nil {
		return true
	}

	return false
}

// SetInfo gets a reference to the given []V1alpha1Info and assigns it to the Info field.
func (o *V1alpha1Operation) SetInfo(v []V1alpha1Info) {
	o.Info = &v
}

// GetInitiatedBy returns the InitiatedBy field value if set, zero value otherwise.
func (o *V1alpha1Operation) GetInitiatedBy() V1alpha1OperationInitiator {
	if o == nil || o.InitiatedBy == nil {
		var ret V1alpha1OperationInitiator
		return ret
	}
	return *o.InitiatedBy
}

// GetInitiatedByOk returns a tuple with the InitiatedBy field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1Operation) GetInitiatedByOk() (*V1alpha1OperationInitiator, bool) {
	if o == nil || o.InitiatedBy == nil {
		return nil, false
	}
	return o.InitiatedBy, true
}

// HasInitiatedBy returns a boolean if a field has been set.
func (o *V1alpha1Operation) HasInitiatedBy() bool {
	if o != nil && o.InitiatedBy != nil {
		return true
	}

	return false
}

// SetInitiatedBy gets a reference to the given V1alpha1OperationInitiator and assigns it to the InitiatedBy field.
func (o *V1alpha1Operation) SetInitiatedBy(v V1alpha1OperationInitiator) {
	o.InitiatedBy = &v
}

// GetRetry returns the Retry field value if set, zero value otherwise.
func (o *V1alpha1Operation) GetRetry() V1alpha1RetryStrategy {
	if o == nil || o.Retry == nil {
		var ret V1alpha1RetryStrategy
		return ret
	}
	return *o.Retry
}

// GetRetryOk returns a tuple with the Retry field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1Operation) GetRetryOk() (*V1alpha1RetryStrategy, bool) {
	if o == nil || o.Retry == nil {
		return nil, false
	}
	return o.Retry, true
}

// HasRetry returns a boolean if a field has been set.
func (o *V1alpha1Operation) HasRetry() bool {
	if o != nil && o.Retry != nil {
		return true
	}

	return false
}

// SetRetry gets a reference to the given V1alpha1RetryStrategy and assigns it to the Retry field.
func (o *V1alpha1Operation) SetRetry(v V1alpha1RetryStrategy) {
	o.Retry = &v
}

// GetSync returns the Sync field value if set, zero value otherwise.
func (o *V1alpha1Operation) GetSync() V1alpha1SyncOperation {
	if o == nil || o.Sync == nil {
		var ret V1alpha1SyncOperation
		return ret
	}
	return *o.Sync
}

// GetSyncOk returns a tuple with the Sync field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1Operation) GetSyncOk() (*V1alpha1SyncOperation, bool) {
	if o == nil || o.Sync == nil {
		return nil, false
	}
	return o.Sync, true
}

// HasSync returns a boolean if a field has been set.
func (o *V1alpha1Operation) HasSync() bool {
	if o != nil && o.Sync != nil {
		return true
	}

	return false
}

// SetSync gets a reference to the given V1alpha1SyncOperation and assigns it to the Sync field.
func (o *V1alpha1Operation) SetSync(v V1alpha1SyncOperation) {
	o.Sync = &v
}

func (o V1alpha1Operation) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Info != nil {
		toSerialize["info"] = o.Info
	}
	if o.InitiatedBy != nil {
		toSerialize["initiatedBy"] = o.InitiatedBy
	}
	if o.Retry != nil {
		toSerialize["retry"] = o.Retry
	}
	if o.Sync != nil {
		toSerialize["sync"] = o.Sync
	}
	return json.Marshal(toSerialize)
}

type NullableV1alpha1Operation struct {
	value *V1alpha1Operation
	isSet bool
}

func (v NullableV1alpha1Operation) Get() *V1alpha1Operation {
	return v.value
}

func (v *NullableV1alpha1Operation) Set(val *V1alpha1Operation) {
	v.value = val
	v.isSet = true
}

func (v NullableV1alpha1Operation) IsSet() bool {
	return v.isSet
}

func (v *NullableV1alpha1Operation) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1alpha1Operation(val *V1alpha1Operation) *NullableV1alpha1Operation {
	return &NullableV1alpha1Operation{value: val, isSet: true}
}

func (v NullableV1alpha1Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1alpha1Operation) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}