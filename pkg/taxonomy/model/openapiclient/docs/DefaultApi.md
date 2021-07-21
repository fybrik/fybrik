# \DefaultApi

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetPoliciesDecisions**](DefaultApi.md#GetPoliciesDecisions) | **Get** /getPoliciesDecisions | getPoliciesDecisions



## GetPoliciesDecisions

> PolicyManagerResponse GetPoliciesDecisions(ctx).Input(input).Creds(creds).Execute()

getPoliciesDecisions



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
    input := TODO // PolicyManagerRequest | Policy Manager Request Object
    creds := "creds_example" // string | credentials to wkc connector

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.GetPoliciesDecisions(context.Background()).Input(input).Creds(creds).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.GetPoliciesDecisions``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPoliciesDecisions`: PolicyManagerResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.GetPoliciesDecisions`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetPoliciesDecisionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **input** | [**PolicyManagerRequest**](PolicyManagerRequest.md) | Policy Manager Request Object | 
 **creds** | **string** | credentials to wkc connector | 

### Return type

[**PolicyManagerResponse**](PolicyManagerResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

