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

// NotificationTriggerList struct for NotificationTriggerList
type NotificationTriggerList struct {
	Items *[]NotificationTrigger `json:"items,omitempty"`
}

// NewNotificationTriggerList instantiates a new NotificationTriggerList object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNotificationTriggerList() *NotificationTriggerList {
	this := NotificationTriggerList{}
	return &this
}

// NewNotificationTriggerListWithDefaults instantiates a new NotificationTriggerList object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNotificationTriggerListWithDefaults() *NotificationTriggerList {
	this := NotificationTriggerList{}
	return &this
}

// GetItems returns the Items field value if set, zero value otherwise.
func (o *NotificationTriggerList) GetItems() []NotificationTrigger {
	if o == nil || o.Items == nil {
		var ret []NotificationTrigger
		return ret
	}
	return *o.Items
}

// GetItemsOk returns a tuple with the Items field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *NotificationTriggerList) GetItemsOk() (*[]NotificationTrigger, bool) {
	if o == nil || o.Items == nil {
		return nil, false
	}
	return o.Items, true
}

// HasItems returns a boolean if a field has been set.
func (o *NotificationTriggerList) HasItems() bool {
	if o != nil && o.Items != nil {
		return true
	}

	return false
}

// SetItems gets a reference to the given []NotificationTrigger and assigns it to the Items field.
func (o *NotificationTriggerList) SetItems(v []NotificationTrigger) {
	o.Items = &v
}

func (o NotificationTriggerList) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Items != nil {
		toSerialize["items"] = o.Items
	}
	return json.Marshal(toSerialize)
}

type NullableNotificationTriggerList struct {
	value *NotificationTriggerList
	isSet bool
}

func (v NullableNotificationTriggerList) Get() *NotificationTriggerList {
	return v.value
}

func (v *NullableNotificationTriggerList) Set(val *NotificationTriggerList) {
	v.value = val
	v.isSet = true
}

func (v NullableNotificationTriggerList) IsSet() bool {
	return v.isSet
}

func (v *NullableNotificationTriggerList) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNotificationTriggerList(val *NotificationTriggerList) *NullableNotificationTriggerList {
	return &NullableNotificationTriggerList{value: val, isSet: true}
}

func (v NullableNotificationTriggerList) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNotificationTriggerList) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
