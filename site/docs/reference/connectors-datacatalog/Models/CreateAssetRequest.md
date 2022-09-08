# CreateAssetRequest

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**credentials** | String | The vault plugin path where the destination data credentials will be stored as kubernetes secrets | [optional] [default: null]
**destinationAssetID** | String | Asset ID to be used for the created asset | [optional] [default: null]
**destinationCatalogID** | String | The destination catalog id in which the new asset will be created based on the information provided in ResourceMetadata and ResourceDetails field | [default: null]
**details** | [ResourceDetails](../Models/ResourceDetails.md) |  | [default: null]
**resourceMetadata** | [ResourceMetadata](../Models/ResourceMetadata.md) |  | [default: null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to API-Specification]](../README.md)

