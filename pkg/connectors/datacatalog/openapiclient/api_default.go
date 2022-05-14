/*
Data Catalog Service - Asset Details

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapiclient

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Linger please
var (
	_ context.Context
)

// DefaultApiService DefaultApi service
type DefaultApiService service

type ApiCreateAssetRequest struct {
	ctx                          context.Context
	ApiService                   *DefaultApiService
	xRequestDatacatalogWriteCred *string
	createAssetRequest           *CreateAssetRequest
}

// This header carries credential information related to accessing the relevant destination catalog.
func (r ApiCreateAssetRequest) XRequestDatacatalogWriteCred(xRequestDatacatalogWriteCred string) ApiCreateAssetRequest {
	r.xRequestDatacatalogWriteCred = &xRequestDatacatalogWriteCred
	return r
}

// Write Asset Request
func (r ApiCreateAssetRequest) CreateAssetRequest(createAssetRequest CreateAssetRequest) ApiCreateAssetRequest {
	r.createAssetRequest = &createAssetRequest
	return r
}

func (r ApiCreateAssetRequest) Execute() (*CreateAssetResponse, *http.Response, error) {
	return r.ApiService.CreateAssetExecute(r)
}

/*
CreateAsset This REST API writes data asset information to the data catalog configured in fybrik

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiCreateAssetRequest
*/
func (a *DefaultApiService) CreateAsset(ctx context.Context) ApiCreateAssetRequest {
	return ApiCreateAssetRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//  @return CreateAssetResponse
func (a *DefaultApiService) CreateAssetExecute(r ApiCreateAssetRequest) (*CreateAssetResponse, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *CreateAssetResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "DefaultApiService.CreateAsset")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/createAsset"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.xRequestDatacatalogWriteCred == nil {
		return localVarReturnValue, nil, reportError("xRequestDatacatalogWriteCred is required and must be specified")
	}
	if r.createAssetRequest == nil {
		return localVarReturnValue, nil, reportError("createAssetRequest is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	localVarHeaderParams["X-Request-Datacatalog-Write-Cred"] = parameterToString(*r.xRequestDatacatalogWriteCred, "")
	// body params
	localVarPostBody = r.createAssetRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiDeleteAssetRequest struct {
	ctx                     context.Context
	ApiService              *DefaultApiService
	xRequestDatacatalogCred *string
	deleteAssetRequest      *DeleteAssetRequest
}

// This header carries credential information related to relevant catalog from which the asset information needs to be retrieved.
func (r ApiDeleteAssetRequest) XRequestDatacatalogCred(xRequestDatacatalogCred string) ApiDeleteAssetRequest {
	r.xRequestDatacatalogCred = &xRequestDatacatalogCred
	return r
}

// Delete Asset Request
func (r ApiDeleteAssetRequest) DeleteAssetRequest(deleteAssetRequest DeleteAssetRequest) ApiDeleteAssetRequest {
	r.deleteAssetRequest = &deleteAssetRequest
	return r
}

func (r ApiDeleteAssetRequest) Execute() (*DeleteAssetResponse, *http.Response, error) {
	return r.ApiService.DeleteAssetExecute(r)
}

/*
DeleteAsset This REST API deletes data asset

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiDeleteAssetRequest
*/
func (a *DefaultApiService) DeleteAsset(ctx context.Context) ApiDeleteAssetRequest {
	return ApiDeleteAssetRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//  @return DeleteAssetResponse
func (a *DefaultApiService) DeleteAssetExecute(r ApiDeleteAssetRequest) (*DeleteAssetResponse, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodDelete
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *DeleteAssetResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "DefaultApiService.DeleteAsset")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/deleteAsset"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.xRequestDatacatalogCred == nil {
		return localVarReturnValue, nil, reportError("xRequestDatacatalogCred is required and must be specified")
	}
	if r.deleteAssetRequest == nil {
		return localVarReturnValue, nil, reportError("deleteAssetRequest is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	localVarHeaderParams["X-Request-Datacatalog-Cred"] = parameterToString(*r.xRequestDatacatalogCred, "")
	// body params
	localVarPostBody = r.deleteAssetRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetAssetInfoRequest struct {
	ctx                     context.Context
	ApiService              *DefaultApiService
	xRequestDatacatalogCred *string
	getAssetRequest         *GetAssetRequest
}

// This header carries credential information related to relevant catalog from which the asset information needs to be retrieved.
func (r ApiGetAssetInfoRequest) XRequestDatacatalogCred(xRequestDatacatalogCred string) ApiGetAssetInfoRequest {
	r.xRequestDatacatalogCred = &xRequestDatacatalogCred
	return r
}

// Data Catalog Request Object.
func (r ApiGetAssetInfoRequest) GetAssetRequest(getAssetRequest GetAssetRequest) ApiGetAssetInfoRequest {
	r.getAssetRequest = &getAssetRequest
	return r
}

func (r ApiGetAssetInfoRequest) Execute() (*GetAssetResponse, *http.Response, error) {
	return r.ApiService.GetAssetInfoExecute(r)
}

/*
GetAssetInfo This REST API gets data asset information from the data catalog configured in fybrik for the data sets indicated in FybrikApplication yaml

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiGetAssetInfoRequest
*/
func (a *DefaultApiService) GetAssetInfo(ctx context.Context) ApiGetAssetInfoRequest {
	return ApiGetAssetInfoRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//  @return GetAssetResponse
func (a *DefaultApiService) GetAssetInfoExecute(r ApiGetAssetInfoRequest) (*GetAssetResponse, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *GetAssetResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "DefaultApiService.GetAssetInfo")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/getAssetInfo"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.xRequestDatacatalogCred == nil {
		return localVarReturnValue, nil, reportError("xRequestDatacatalogCred is required and must be specified")
	}
	if r.getAssetRequest == nil {
		return localVarReturnValue, nil, reportError("getAssetRequest is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	localVarHeaderParams["X-Request-Datacatalog-Cred"] = parameterToString(*r.xRequestDatacatalogCred, "")
	// body params
	localVarPostBody = r.getAssetRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiUpdateAssetRequest struct {
	ctx                           context.Context
	ApiService                    *DefaultApiService
	xRequestDatacatalogUpdateCred *string
	updateAssetRequest            *UpdateAssetRequest
}

// This header carries credential information related to accessing the relevant destination catalog.
func (r ApiUpdateAssetRequest) XRequestDatacatalogUpdateCred(xRequestDatacatalogUpdateCred string) ApiUpdateAssetRequest {
	r.xRequestDatacatalogUpdateCred = &xRequestDatacatalogUpdateCred
	return r
}

// Update Asset Request
func (r ApiUpdateAssetRequest) UpdateAssetRequest(updateAssetRequest UpdateAssetRequest) ApiUpdateAssetRequest {
	r.updateAssetRequest = &updateAssetRequest
	return r
}

func (r ApiUpdateAssetRequest) Execute() (*UpdateAssetResponse, *http.Response, error) {
	return r.ApiService.UpdateAssetExecute(r)
}

/*
UpdateAsset This REST API updates data asset information to the data catalog configured in fybrik

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiUpdateAssetRequest
*/
func (a *DefaultApiService) UpdateAsset(ctx context.Context) ApiUpdateAssetRequest {
	return ApiUpdateAssetRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//  @return UpdateAssetResponse
func (a *DefaultApiService) UpdateAssetExecute(r ApiUpdateAssetRequest) (*UpdateAssetResponse, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodPut
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *UpdateAssetResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "DefaultApiService.UpdateAsset")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/updateAsset"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.xRequestDatacatalogUpdateCred == nil {
		return localVarReturnValue, nil, reportError("xRequestDatacatalogUpdateCred is required and must be specified")
	}
	if r.updateAssetRequest == nil {
		return localVarReturnValue, nil, reportError("updateAssetRequest is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	localVarHeaderParams["X-Request-Datacatalog-Update-Cred"] = parameterToString(*r.xRequestDatacatalogUpdateCred, "")
	// body params
	localVarPostBody = r.updateAssetRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
