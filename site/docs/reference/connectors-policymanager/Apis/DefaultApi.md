# DefaultApi

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getPoliciesDecisions**](DefaultApi.md#getPoliciesDecisions) | **POST** /getPoliciesDecisions | This REST API gets data governance decisions for the data sets indicated in FybrikApplication yaml based on the context indicated


<a name="getPoliciesDecisions"></a>
## **getPoliciesDecisions**
> GetPolicyDecisionsResponse getPoliciesDecisions(X-Request-CredGetPolicyDecisionsRequest)

This REST API gets data governance decisions for the data sets indicated in FybrikApplication yaml based on the context indicated


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**X-Request-Cred**|**String**|  | [default to null]
**GetPolicyDecisionsRequest**|[**GetPolicyDecisionsRequest**](../Models/GetPolicyDecisionsRequest.md)| Policy Manager Request Object. |

### Return type


[**GetPolicyDecisionsResponse**](../Models/GetPolicyDecisionsResponse.md)



### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

 [[Back to API-Specification]](../README.md) 

