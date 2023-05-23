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

// V1alpha1ClusterConfig ClusterConfig is the configuration attributes. This structure is subset of the go-client rest.Config with annotations added for marshalling.
type V1alpha1ClusterConfig struct {
	AwsAuthConfig *V1alpha1AWSAuthConfig `json:"awsAuthConfig,omitempty"`
	// Server requires Bearer authentication. This client will not attempt to use refresh tokens for an OAuth2 flow. TODO: demonstrate an OAuth2 compatible client.
	BearerToken        *string                     `json:"bearerToken,omitempty"`
	ExecProviderConfig *V1alpha1ExecProviderConfig `json:"execProviderConfig,omitempty"`
	Password           *string                     `json:"password,omitempty"`
	TlsClientConfig    *V1alpha1TLSClientConfig    `json:"tlsClientConfig,omitempty"`
	Username           *string                     `json:"username,omitempty"`
}

// NewV1alpha1ClusterConfig instantiates a new V1alpha1ClusterConfig object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1alpha1ClusterConfig() *V1alpha1ClusterConfig {
	this := V1alpha1ClusterConfig{}
	return &this
}

// NewV1alpha1ClusterConfigWithDefaults instantiates a new V1alpha1ClusterConfig object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1alpha1ClusterConfigWithDefaults() *V1alpha1ClusterConfig {
	this := V1alpha1ClusterConfig{}
	return &this
}

// GetAwsAuthConfig returns the AwsAuthConfig field value if set, zero value otherwise.
func (o *V1alpha1ClusterConfig) GetAwsAuthConfig() V1alpha1AWSAuthConfig {
	if o == nil || o.AwsAuthConfig == nil {
		var ret V1alpha1AWSAuthConfig
		return ret
	}
	return *o.AwsAuthConfig
}

// GetAwsAuthConfigOk returns a tuple with the AwsAuthConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ClusterConfig) GetAwsAuthConfigOk() (*V1alpha1AWSAuthConfig, bool) {
	if o == nil || o.AwsAuthConfig == nil {
		return nil, false
	}
	return o.AwsAuthConfig, true
}

// HasAwsAuthConfig returns a boolean if a field has been set.
func (o *V1alpha1ClusterConfig) HasAwsAuthConfig() bool {
	if o != nil && o.AwsAuthConfig != nil {
		return true
	}

	return false
}

// SetAwsAuthConfig gets a reference to the given V1alpha1AWSAuthConfig and assigns it to the AwsAuthConfig field.
func (o *V1alpha1ClusterConfig) SetAwsAuthConfig(v V1alpha1AWSAuthConfig) {
	o.AwsAuthConfig = &v
}

// GetBearerToken returns the BearerToken field value if set, zero value otherwise.
func (o *V1alpha1ClusterConfig) GetBearerToken() string {
	if o == nil || o.BearerToken == nil {
		var ret string
		return ret
	}
	return *o.BearerToken
}

// GetBearerTokenOk returns a tuple with the BearerToken field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ClusterConfig) GetBearerTokenOk() (*string, bool) {
	if o == nil || o.BearerToken == nil {
		return nil, false
	}
	return o.BearerToken, true
}

// HasBearerToken returns a boolean if a field has been set.
func (o *V1alpha1ClusterConfig) HasBearerToken() bool {
	if o != nil && o.BearerToken != nil {
		return true
	}

	return false
}

// SetBearerToken gets a reference to the given string and assigns it to the BearerToken field.
func (o *V1alpha1ClusterConfig) SetBearerToken(v string) {
	o.BearerToken = &v
}

// GetExecProviderConfig returns the ExecProviderConfig field value if set, zero value otherwise.
func (o *V1alpha1ClusterConfig) GetExecProviderConfig() V1alpha1ExecProviderConfig {
	if o == nil || o.ExecProviderConfig == nil {
		var ret V1alpha1ExecProviderConfig
		return ret
	}
	return *o.ExecProviderConfig
}

// GetExecProviderConfigOk returns a tuple with the ExecProviderConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ClusterConfig) GetExecProviderConfigOk() (*V1alpha1ExecProviderConfig, bool) {
	if o == nil || o.ExecProviderConfig == nil {
		return nil, false
	}
	return o.ExecProviderConfig, true
}

// HasExecProviderConfig returns a boolean if a field has been set.
func (o *V1alpha1ClusterConfig) HasExecProviderConfig() bool {
	if o != nil && o.ExecProviderConfig != nil {
		return true
	}

	return false
}

// SetExecProviderConfig gets a reference to the given V1alpha1ExecProviderConfig and assigns it to the ExecProviderConfig field.
func (o *V1alpha1ClusterConfig) SetExecProviderConfig(v V1alpha1ExecProviderConfig) {
	o.ExecProviderConfig = &v
}

// GetPassword returns the Password field value if set, zero value otherwise.
func (o *V1alpha1ClusterConfig) GetPassword() string {
	if o == nil || o.Password == nil {
		var ret string
		return ret
	}
	return *o.Password
}

// GetPasswordOk returns a tuple with the Password field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ClusterConfig) GetPasswordOk() (*string, bool) {
	if o == nil || o.Password == nil {
		return nil, false
	}
	return o.Password, true
}

// HasPassword returns a boolean if a field has been set.
func (o *V1alpha1ClusterConfig) HasPassword() bool {
	if o != nil && o.Password != nil {
		return true
	}

	return false
}

// SetPassword gets a reference to the given string and assigns it to the Password field.
func (o *V1alpha1ClusterConfig) SetPassword(v string) {
	o.Password = &v
}

// GetTlsClientConfig returns the TlsClientConfig field value if set, zero value otherwise.
func (o *V1alpha1ClusterConfig) GetTlsClientConfig() V1alpha1TLSClientConfig {
	if o == nil || o.TlsClientConfig == nil {
		var ret V1alpha1TLSClientConfig
		return ret
	}
	return *o.TlsClientConfig
}

// GetTlsClientConfigOk returns a tuple with the TlsClientConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ClusterConfig) GetTlsClientConfigOk() (*V1alpha1TLSClientConfig, bool) {
	if o == nil || o.TlsClientConfig == nil {
		return nil, false
	}
	return o.TlsClientConfig, true
}

// HasTlsClientConfig returns a boolean if a field has been set.
func (o *V1alpha1ClusterConfig) HasTlsClientConfig() bool {
	if o != nil && o.TlsClientConfig != nil {
		return true
	}

	return false
}

// SetTlsClientConfig gets a reference to the given V1alpha1TLSClientConfig and assigns it to the TlsClientConfig field.
func (o *V1alpha1ClusterConfig) SetTlsClientConfig(v V1alpha1TLSClientConfig) {
	o.TlsClientConfig = &v
}

// GetUsername returns the Username field value if set, zero value otherwise.
func (o *V1alpha1ClusterConfig) GetUsername() string {
	if o == nil || o.Username == nil {
		var ret string
		return ret
	}
	return *o.Username
}

// GetUsernameOk returns a tuple with the Username field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1alpha1ClusterConfig) GetUsernameOk() (*string, bool) {
	if o == nil || o.Username == nil {
		return nil, false
	}
	return o.Username, true
}

// HasUsername returns a boolean if a field has been set.
func (o *V1alpha1ClusterConfig) HasUsername() bool {
	if o != nil && o.Username != nil {
		return true
	}

	return false
}

// SetUsername gets a reference to the given string and assigns it to the Username field.
func (o *V1alpha1ClusterConfig) SetUsername(v string) {
	o.Username = &v
}

func (o V1alpha1ClusterConfig) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.AwsAuthConfig != nil {
		toSerialize["awsAuthConfig"] = o.AwsAuthConfig
	}
	if o.BearerToken != nil {
		toSerialize["bearerToken"] = o.BearerToken
	}
	if o.ExecProviderConfig != nil {
		toSerialize["execProviderConfig"] = o.ExecProviderConfig
	}
	if o.Password != nil {
		toSerialize["password"] = o.Password
	}
	if o.TlsClientConfig != nil {
		toSerialize["tlsClientConfig"] = o.TlsClientConfig
	}
	if o.Username != nil {
		toSerialize["username"] = o.Username
	}
	return json.Marshal(toSerialize)
}

type NullableV1alpha1ClusterConfig struct {
	value *V1alpha1ClusterConfig
	isSet bool
}

func (v NullableV1alpha1ClusterConfig) Get() *V1alpha1ClusterConfig {
	return v.value
}

func (v *NullableV1alpha1ClusterConfig) Set(val *V1alpha1ClusterConfig) {
	v.value = val
	v.isSet = true
}

func (v NullableV1alpha1ClusterConfig) IsSet() bool {
	return v.isSet
}

func (v *NullableV1alpha1ClusterConfig) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1alpha1ClusterConfig(val *V1alpha1ClusterConfig) *NullableV1alpha1ClusterConfig {
	return &NullableV1alpha1ClusterConfig{value: val, isSet: true}
}

func (v NullableV1alpha1ClusterConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1alpha1ClusterConfig) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}