# ResourceColumns

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Tags** | Pointer to **map[string]interface{}** |  | [optional] 

## Methods

### NewResourceColumns

`func NewResourceColumns(name string, ) *ResourceColumns`

NewResourceColumns instantiates a new ResourceColumns object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceColumnsWithDefaults

`func NewResourceColumnsWithDefaults() *ResourceColumns`

NewResourceColumnsWithDefaults instantiates a new ResourceColumns object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ResourceColumns) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ResourceColumns) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ResourceColumns) SetName(v string)`

SetName sets Name field to given value.


### GetTags

`func (o *ResourceColumns) GetTags() map[string]interface{}`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *ResourceColumns) GetTagsOk() (*map[string]interface{}, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *ResourceColumns) SetTags(v map[string]interface{})`

SetTags sets Tags field to given value.

### HasTags

`func (o *ResourceColumns) HasTags() bool`

HasTags returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


