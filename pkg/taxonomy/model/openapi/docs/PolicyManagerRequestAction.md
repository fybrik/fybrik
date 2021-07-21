# PolicyManagerRequestAction

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActionType** | Pointer to [**ActionType**](ActionType.md) |  | [optional] 
**ProcessingLocation** | Pointer to **string** |  | [optional] 

## Methods

### NewPolicyManagerRequestAction

`func NewPolicyManagerRequestAction() *PolicyManagerRequestAction`

NewPolicyManagerRequestAction instantiates a new PolicyManagerRequestAction object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyManagerRequestActionWithDefaults

`func NewPolicyManagerRequestActionWithDefaults() *PolicyManagerRequestAction`

NewPolicyManagerRequestActionWithDefaults instantiates a new PolicyManagerRequestAction object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActionType

`func (o *PolicyManagerRequestAction) GetActionType() ActionType`

GetActionType returns the ActionType field if non-nil, zero value otherwise.

### GetActionTypeOk

`func (o *PolicyManagerRequestAction) GetActionTypeOk() (*ActionType, bool)`

GetActionTypeOk returns a tuple with the ActionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionType

`func (o *PolicyManagerRequestAction) SetActionType(v ActionType)`

SetActionType sets ActionType field to given value.

### HasActionType

`func (o *PolicyManagerRequestAction) HasActionType() bool`

HasActionType returns a boolean if a field has been set.

### GetProcessingLocation

`func (o *PolicyManagerRequestAction) GetProcessingLocation() string`

GetProcessingLocation returns the ProcessingLocation field if non-nil, zero value otherwise.

### GetProcessingLocationOk

`func (o *PolicyManagerRequestAction) GetProcessingLocationOk() (*string, bool)`

GetProcessingLocationOk returns a tuple with the ProcessingLocation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProcessingLocation

`func (o *PolicyManagerRequestAction) SetProcessingLocation(v string)`

SetProcessingLocation sets ProcessingLocation field to given value.

### HasProcessingLocation

`func (o *PolicyManagerRequestAction) HasProcessingLocation() bool`

HasProcessingLocation returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


