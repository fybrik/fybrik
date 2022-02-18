# \DefaultApi

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateAssetInfo**](DefaultApi.md#CreateAssetInfo) | **Post** /createAssetInfo | This REST API writes data asset information to the data catalog configured in fybrik
[**GetAssetInfo**](DefaultApi.md#GetAssetInfo) | **Post** /getAssetInfo | This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml



## CreateAssetInfo

> CreateAssetResponse CreateAssetInfo(ctx).XRequestDatacatalogWriteCred(xRequestDatacatalogWriteCred).CreateAssetRequest(createAssetRequest).Execute()

This REST API writes data asset information to the data catalog configured in fybrik

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    xRequestDatacatalogWriteCred := "xRequestDatacatalogWriteCred_example" // string | 
    createAssetRequest := TODO // CreateAssetRequest | Write Asset Request

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.CreateAssetInfo(context.Background()).XRequestDatacatalogWriteCred(xRequestDatacatalogWriteCred).CreateAssetRequest(createAssetRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.CreateAssetInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateAssetInfo`: CreateAssetResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.CreateAssetInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateAssetInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xRequestDatacatalogWriteCred** | **string** |  | 
 **createAssetRequest** | [**CreateAssetRequest**](CreateAssetRequest.md) | Write Asset Request | 

### Return type

[**CreateAssetResponse**](CreateAssetResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAssetInfo

> GetAssetResponse GetAssetInfo(ctx).XRequestDatacatalogCred(xRequestDatacatalogCred).GetAssetRequest(getAssetRequest).Execute()

This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    xRequestDatacatalogCred := "xRequestDatacatalogCred_example" // string | 
    getAssetRequest := TODO // GetAssetRequest | Data Catalog Request Object.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.GetAssetInfo(context.Background()).XRequestDatacatalogCred(xRequestDatacatalogCred).GetAssetRequest(getAssetRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.GetAssetInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAssetInfo`: GetAssetResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.GetAssetInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAssetInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xRequestDatacatalogCred** | **string** |  | 
 **getAssetRequest** | [**GetAssetRequest**](GetAssetRequest.md) | Data Catalog Request Object. | 

### Return type

[**GetAssetResponse**](GetAssetResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

