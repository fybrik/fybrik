# DefaultApi

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getAssetInfoPost**](DefaultApi.md#getAssetInfoPost) | **POST** /getAssetInfo | getAssetInfo


<a name="getAssetInfoPost"></a>
# **getAssetInfoPost**
> GetAssetResponse getAssetInfoPost(X-Request-Datacatalog-Cred, GetAssetRequest)

getAssetInfo

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **X-Request-Datacatalog-Cred** | **String**|  | [default to null]
 **GetAssetRequest** | [**GetAssetRequest**](../Models/GetAssetRequest.md)| Data Catalog Request Object. |

### Return type

[**GetAssetResponse**](../Models/GetAssetResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

