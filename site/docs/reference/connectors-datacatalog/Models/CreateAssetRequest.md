# CreateAssetRequest

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**credentials** | String | This optional field has the vault plugin path where the destination data credentials will be stored as kubernetes secrets | [optional] [default: null]
**destinationAssetID** | String | This is an optional field provided to give information about the asset id to be used for the created asset. | [optional] [default: null]
**destinationCatalogID** | String | This has the information about the destination catalog id that new asset that will be created with the information provided in ResourceMetadata and Details field will be stored. | [default: null]
**resourceDetails** | [ResourceDetails](../ResourceDetails) |  | [default: null]
**resourceMetadata** | [ResourceMetadata](../ResourceMetadata) |  | [default: null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to API-Specification]](../README.md)

