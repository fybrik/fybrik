# PolicyManagerRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Context** | Pointer to **map[string]interface{}** |  | [optional] 
**Action** | [**PolicyManagerRequestAction**](PolicyManagerRequestAction.md) |  | 
**Resource** | [**Resource**](Resource.md) |  | 

## Methods

### NewPolicyManagerRequest

`func NewPolicyManagerRequest(action PolicyManagerRequestAction, resource Resource, ) *PolicyManagerRequest`

NewPolicyManagerRequest instantiates a new PolicyManagerRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyManagerRequestWithDefaults

`func NewPolicyManagerRequestWithDefaults() *PolicyManagerRequest`

NewPolicyManagerRequestWithDefaults instantiates a new PolicyManagerRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContext

`func (o *PolicyManagerRequest) GetContext() map[string]interface{}`

GetContext returns the Context field if non-nil, zero value otherwise.

### GetContextOk

`func (o *PolicyManagerRequest) GetContextOk() (*map[string]interface{}, bool)`

GetContextOk returns a tuple with the Context field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContext

`func (o *PolicyManagerRequest) SetContext(v map[string]interface{})`

SetContext sets Context field to given value.

### HasContext

`func (o *PolicyManagerRequest) HasContext() bool`

HasContext returns a boolean if a field has been set.

### GetAction

`func (o *PolicyManagerRequest) GetAction() PolicyManagerRequestAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *PolicyManagerRequest) GetActionOk() (*PolicyManagerRequestAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *PolicyManagerRequest) SetAction(v PolicyManagerRequestAction)`

SetAction sets Action field to given value.


### GetResource

`func (o *PolicyManagerRequest) GetResource() Resource`

GetResource returns the Resource field if non-nil, zero value otherwise.

### GetResourceOk

`func (o *PolicyManagerRequest) GetResourceOk() (*Resource, bool)`

GetResourceOk returns a tuple with the Resource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResource

`func (o *PolicyManagerRequest) SetResource(v Resource)`

SetResource sets Resource field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


