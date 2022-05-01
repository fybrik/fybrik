# DefaultApi

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**createAsset**](DefaultApi.md#createAsset) | **POST** /createAsset | This REST API writes data asset information to the data catalog configured in fybrik
[**deleteAsset**](DefaultApi.md#deleteAsset) | **DELETE** /deleteAsset | This REST API deletes data asset
[**getAssetInfo**](DefaultApi.md#getAssetInfo) | **POST** /getAssetInfo | This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml


<a name="createAsset"></a>
## **createAsset**
> CreateAssetResponse createAsset(X-Request-Datacatalog-Write-CredCreateAssetRequest)

This REST API writes data asset information to the data catalog configured in fybrik


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**X-Request-Datacatalog-Write-Cred**|**String**| This header carries credential information related to accessing the relevant destination catalog. | [default to null]
**CreateAssetRequest**|**CreateAssetRequest**| Write Asset Request |

### Return type


[**CreateAssetResponse**](../Models/CreateAssetResponse.md)



### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

 [[Back to API-Specification]](../README.md) 

<a name="deleteAsset"></a>
## **deleteAsset**
> DeleteAssetResponse deleteAsset(X-Request-Datacatalog-CredDeleteAssetRequest)

This REST API deletes data asset


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**X-Request-Datacatalog-Cred**|**String**| This header carries credential information related to relevant catalog from which the asset information needs to be retrieved. | [default to null]
**DeleteAssetRequest**|**DeleteAssetRequest**| Delete Asset Request |

### Return type


[**DeleteAssetResponse**](../Models/DeleteAssetResponse.md)



### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

 [[Back to API-Specification]](../README.md) 

<a name="getAssetInfo"></a>
## **getAssetInfo**
> GetAssetResponse getAssetInfo(X-Request-Datacatalog-CredGetAssetRequest)

This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**X-Request-Datacatalog-Cred**|**String**| This header carries credential information related to relevant catalog from which the asset information needs to be retrieved. | [default to null]
**GetAssetRequest**|**GetAssetRequest**| Data Catalog Request Object. |

### Return type


[**GetAssetResponse**](../Models/GetAssetResponse.md)



### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

 [[Back to API-Specification]](../README.md) 

