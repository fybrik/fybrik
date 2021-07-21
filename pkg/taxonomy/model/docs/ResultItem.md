# ResultItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Policy** | **string** | The policy on which the decision was based. | 
**Action** | [**Action**](Action.md) |  | 

## Methods

### NewResultItem

`func NewResultItem(policy string, action Action, ) *ResultItem`

NewResultItem instantiates a new ResultItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResultItemWithDefaults

`func NewResultItemWithDefaults() *ResultItem`

NewResultItemWithDefaults instantiates a new ResultItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPolicy

`func (o *ResultItem) GetPolicy() string`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *ResultItem) GetPolicyOk() (*string, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *ResultItem) SetPolicy(v string)`

SetPolicy sets Policy field to given value.


### GetAction

`func (o *ResultItem) GetAction() Action`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *ResultItem) GetActionOk() (*Action, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *ResultItem) SetAction(v Action)`

SetAction sets Action field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


