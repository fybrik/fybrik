# DefaultApi

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getAssetInfo**](DefaultApi.md#getAssetInfo) | **POST** /getAssetInfo | This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml


<a name="getAssetInfo"></a>
## **getAssetInfo**
> GetAssetResponse getAssetInfo(X-Request-Datacatalog-CredGetAssetRequest)

This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**X-Request-Datacatalog-Cred**|**String**|  | [default to null]
**GetAssetRequest**|**GetAssetRequest**| Data Catalog Request Object. |

### Return type


[**GetAssetResponse**](../Models/GetAssetResponse.md)



### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

 [[Back to API-Specification]](../README.md) 

