# DefaultApi

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**createAssetInfo**](DefaultApi.md#createAssetInfo) | **POST** /createAssetInfo | This REST API writes data asset information to the data catalog configured in fybrik
[**getAssetInfo**](DefaultApi.md#getAssetInfo) | **POST** /getAssetInfo | This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml


<a name="createAssetInfo"></a>
## **createAssetInfo**
> CreateAssetResponse createAssetInfo(X-Request-Datacatalog-Write-CredCreateAssetRequest)

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

