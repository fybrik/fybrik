# PolicyManagerResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DecisionId** | Pointer to **string** |  | [optional] 
**Result** | [**[]ResultItem**](ResultItem.md) |  | 

## Methods

### NewPolicyManagerResponse

`func NewPolicyManagerResponse(result []ResultItem, ) *PolicyManagerResponse`

NewPolicyManagerResponse instantiates a new PolicyManagerResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyManagerResponseWithDefaults

`func NewPolicyManagerResponseWithDefaults() *PolicyManagerResponse`

NewPolicyManagerResponseWithDefaults instantiates a new PolicyManagerResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDecisionId

`func (o *PolicyManagerResponse) GetDecisionId() string`

GetDecisionId returns the DecisionId field if non-nil, zero value otherwise.

### GetDecisionIdOk

`func (o *PolicyManagerResponse) GetDecisionIdOk() (*string, bool)`

GetDecisionIdOk returns a tuple with the DecisionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDecisionId

`func (o *PolicyManagerResponse) SetDecisionId(v string)`

SetDecisionId sets DecisionId field to given value.

### HasDecisionId

`func (o *PolicyManagerResponse) HasDecisionId() bool`

HasDecisionId returns a boolean if a field has been set.

### GetResult

`func (o *PolicyManagerResponse) GetResult() []ResultItem`

GetResult returns the Result field if non-nil, zero value otherwise.

### GetResultOk

`func (o *PolicyManagerResponse) GetResultOk() (*[]ResultItem, bool)`

GetResultOk returns a tuple with the Result field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResult

`func (o *PolicyManagerResponse) SetResult(v []ResultItem)`

SetResult sets Result field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


