# DefaultApi

All URIs are relative to *https://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**allocateStorage**](DefaultApi.md#allocateStorage) | **POST** /allocateStorage | This REST API allocates storage based on the storage account selected by Fybrik
[**deleteStorage**](DefaultApi.md#deleteStorage) | **DELETE** /deleteStorage | This REST API deletes allocated storage
[**getSupportedConnections**](DefaultApi.md#getSupportedConnections) | **POST** /getSupportedConnections | This REST API returns a list of supported storage types


<a name="allocateStorage"></a>
## **allocateStorage**
> AllocateStorageResponse allocateStorage(AllocateStorageRequest)

This REST API allocates storage based on the storage account selected by Fybrik


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**AllocateStorageRequest**|[**AllocateStorageRequest**](../Models/AllocateStorageRequest.md)| Allocate Storage Request |

### Return type


[**AllocateStorageResponse**](../Models/AllocateStorageResponse.md)



### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

 [[Back to API-Specification]](../README.md) 

<a name="deleteStorage"></a>
## **deleteStorage**
> deleteStorage(DeleteStorageRequest)

This REST API deletes allocated storage


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**DeleteStorageRequest**|[**DeleteStorageRequest**](../Models/DeleteStorageRequest.md)| Delete Storage Request |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

 [[Back to API-Specification]](../README.md) 

<a name="getSupportedConnections"></a>
## **getSupportedConnections**
> GetSupportedConnectionsResponse getSupportedConnections()

This REST API returns a list of supported storage types


### Parameters
This endpoint does not need any parameter.

### Return type


[**GetSupportedConnectionsResponse**](../Models/GetSupportedConnectionsResponse.md)



### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

 [[Back to API-Specification]](../README.md) 

