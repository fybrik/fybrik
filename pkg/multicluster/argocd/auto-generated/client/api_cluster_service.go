/*
Consolidate Services

Description of all APIs

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapiclient

import (
	"bytes"
	_context "context"
	_ioutil "io/ioutil"
	_nethttp "net/http"
	_neturl "net/url"
	"reflect"
	"strings"
)

// Linger please
var (
	_ _context.Context
)

// ClusterServiceApiService ClusterServiceApi service
type ClusterServiceApiService service

type ApiClusterServiceCreateRequest struct {
	ctx        _context.Context
	ApiService *ClusterServiceApiService
	body       *V1alpha1Cluster
	upsert     *bool
}

func (r ApiClusterServiceCreateRequest) Body(body V1alpha1Cluster) ApiClusterServiceCreateRequest {
	r.body = &body
	return r
}
func (r ApiClusterServiceCreateRequest) Upsert(upsert bool) ApiClusterServiceCreateRequest {
	r.upsert = &upsert
	return r
}

func (r ApiClusterServiceCreateRequest) Execute() (V1alpha1Cluster, *_nethttp.Response, error) {
	return r.ApiService.ClusterServiceCreateExecute(r)
}

/*
ClusterServiceCreate Create creates a cluster

	@param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@return ApiClusterServiceCreateRequest
*/
func (a *ClusterServiceApiService) ClusterServiceCreate(ctx _context.Context) ApiClusterServiceCreateRequest {
	return ApiClusterServiceCreateRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//
//	@return V1alpha1Cluster
func (a *ClusterServiceApiService) ClusterServiceCreateExecute(r ApiClusterServiceCreateRequest) (V1alpha1Cluster, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod  = _nethttp.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue V1alpha1Cluster
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ClusterServiceApiService.ClusterServiceCreate")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/clusters"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}
	if r.body == nil {
		return localVarReturnValue, nil, reportError("body is required and must be specified")
	}

	if r.upsert != nil {
		localVarQueryParams.Add("upsert", parameterToString(*r.upsert, ""))
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
	// body params
	localVarPostBody = r.body
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RuntimeError
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiClusterServiceDeleteRequest struct {
	ctx        _context.Context
	ApiService *ClusterServiceApiService
	idValue    string
	server     *string
	name       *string
	idType     *string
}

func (r ApiClusterServiceDeleteRequest) Server(server string) ApiClusterServiceDeleteRequest {
	r.server = &server
	return r
}
func (r ApiClusterServiceDeleteRequest) Name(name string) ApiClusterServiceDeleteRequest {
	r.name = &name
	return r
}

// type is the type of the specified cluster identifier ( \&quot;server\&quot; - default, \&quot;name\&quot; ).
func (r ApiClusterServiceDeleteRequest) IdType(idType string) ApiClusterServiceDeleteRequest {
	r.idType = &idType
	return r
}

func (r ApiClusterServiceDeleteRequest) Execute() (map[string]interface{}, *_nethttp.Response, error) {
	return r.ApiService.ClusterServiceDeleteExecute(r)
}

/*
ClusterServiceDelete Delete deletes a cluster

	@param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param idValue value holds the cluster server URL or cluster name
	@return ApiClusterServiceDeleteRequest
*/
func (a *ClusterServiceApiService) ClusterServiceDelete(ctx _context.Context, idValue string) ApiClusterServiceDeleteRequest {
	return ApiClusterServiceDeleteRequest{
		ApiService: a,
		ctx:        ctx,
		idValue:    idValue,
	}
}

// Execute executes the request
//
//	@return map[string]interface{}
func (a *ClusterServiceApiService) ClusterServiceDeleteExecute(r ApiClusterServiceDeleteRequest) (map[string]interface{}, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod  = _nethttp.MethodDelete
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue map[string]interface{}
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ClusterServiceApiService.ClusterServiceDelete")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/clusters/{id.value}"
	localVarPath = strings.Replace(localVarPath, "{"+"id.value"+"}", _neturl.PathEscape(parameterToString(r.idValue, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	if r.server != nil {
		localVarQueryParams.Add("server", parameterToString(*r.server, ""))
	}
	if r.name != nil {
		localVarQueryParams.Add("name", parameterToString(*r.name, ""))
	}
	if r.idType != nil {
		localVarQueryParams.Add("id.type", parameterToString(*r.idType, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

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
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RuntimeError
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiClusterServiceGetRequest struct {
	ctx        _context.Context
	ApiService *ClusterServiceApiService
	idValue    string
	server     *string
	name       *string
	idType     *string
}

func (r ApiClusterServiceGetRequest) Server(server string) ApiClusterServiceGetRequest {
	r.server = &server
	return r
}
func (r ApiClusterServiceGetRequest) Name(name string) ApiClusterServiceGetRequest {
	r.name = &name
	return r
}

// type is the type of the specified cluster identifier ( \&quot;server\&quot; - default, \&quot;name\&quot; ).
func (r ApiClusterServiceGetRequest) IdType(idType string) ApiClusterServiceGetRequest {
	r.idType = &idType
	return r
}

func (r ApiClusterServiceGetRequest) Execute() (V1alpha1Cluster, *_nethttp.Response, error) {
	return r.ApiService.ClusterServiceGetExecute(r)
}

/*
ClusterServiceGet Get returns a cluster by server address

	@param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param idValue value holds the cluster server URL or cluster name
	@return ApiClusterServiceGetRequest
*/
func (a *ClusterServiceApiService) ClusterServiceGet(ctx _context.Context, idValue string) ApiClusterServiceGetRequest {
	return ApiClusterServiceGetRequest{
		ApiService: a,
		ctx:        ctx,
		idValue:    idValue,
	}
}

// Execute executes the request
//
//	@return V1alpha1Cluster
func (a *ClusterServiceApiService) ClusterServiceGetExecute(r ApiClusterServiceGetRequest) (V1alpha1Cluster, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod  = _nethttp.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue V1alpha1Cluster
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ClusterServiceApiService.ClusterServiceGet")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/clusters/{id.value}"
	localVarPath = strings.Replace(localVarPath, "{"+"id.value"+"}", _neturl.PathEscape(parameterToString(r.idValue, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	if r.server != nil {
		localVarQueryParams.Add("server", parameterToString(*r.server, ""))
	}
	if r.name != nil {
		localVarQueryParams.Add("name", parameterToString(*r.name, ""))
	}
	if r.idType != nil {
		localVarQueryParams.Add("id.type", parameterToString(*r.idType, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

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
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RuntimeError
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiClusterServiceInvalidateCacheRequest struct {
	ctx        _context.Context
	ApiService *ClusterServiceApiService
	idValue    string
}

func (r ApiClusterServiceInvalidateCacheRequest) Execute() (V1alpha1Cluster, *_nethttp.Response, error) {
	return r.ApiService.ClusterServiceInvalidateCacheExecute(r)
}

/*
ClusterServiceInvalidateCache InvalidateCache invalidates cluster cache

	@param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param idValue value holds the cluster server URL or cluster name
	@return ApiClusterServiceInvalidateCacheRequest
*/
func (a *ClusterServiceApiService) ClusterServiceInvalidateCache(ctx _context.Context, idValue string) ApiClusterServiceInvalidateCacheRequest {
	return ApiClusterServiceInvalidateCacheRequest{
		ApiService: a,
		ctx:        ctx,
		idValue:    idValue,
	}
}

// Execute executes the request
//
//	@return V1alpha1Cluster
func (a *ClusterServiceApiService) ClusterServiceInvalidateCacheExecute(r ApiClusterServiceInvalidateCacheRequest) (V1alpha1Cluster, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod  = _nethttp.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue V1alpha1Cluster
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ClusterServiceApiService.ClusterServiceInvalidateCache")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/clusters/{id.value}/invalidate-cache"
	localVarPath = strings.Replace(localVarPath, "{"+"id.value"+"}", _neturl.PathEscape(parameterToString(r.idValue, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

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
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RuntimeError
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiClusterServiceListRequest struct {
	ctx        _context.Context
	ApiService *ClusterServiceApiService
	server     *string
	name       *string
	idType     *string
	idValue    *string
}

func (r ApiClusterServiceListRequest) Server(server string) ApiClusterServiceListRequest {
	r.server = &server
	return r
}
func (r ApiClusterServiceListRequest) Name(name string) ApiClusterServiceListRequest {
	r.name = &name
	return r
}

// type is the type of the specified cluster identifier ( \&quot;server\&quot; - default, \&quot;name\&quot; ).
func (r ApiClusterServiceListRequest) IdType(idType string) ApiClusterServiceListRequest {
	r.idType = &idType
	return r
}

// value holds the cluster server URL or cluster name.
func (r ApiClusterServiceListRequest) IdValue(idValue string) ApiClusterServiceListRequest {
	r.idValue = &idValue
	return r
}

func (r ApiClusterServiceListRequest) Execute() (V1alpha1ClusterList, *_nethttp.Response, error) {
	return r.ApiService.ClusterServiceListExecute(r)
}

/*
ClusterServiceList List returns list of clusters

	@param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@return ApiClusterServiceListRequest
*/
func (a *ClusterServiceApiService) ClusterServiceList(ctx _context.Context) ApiClusterServiceListRequest {
	return ApiClusterServiceListRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//
//	@return V1alpha1ClusterList
func (a *ClusterServiceApiService) ClusterServiceListExecute(r ApiClusterServiceListRequest) (V1alpha1ClusterList, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod  = _nethttp.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue V1alpha1ClusterList
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ClusterServiceApiService.ClusterServiceList")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/clusters"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	if r.server != nil {
		localVarQueryParams.Add("server", parameterToString(*r.server, ""))
	}
	if r.name != nil {
		localVarQueryParams.Add("name", parameterToString(*r.name, ""))
	}
	if r.idType != nil {
		localVarQueryParams.Add("id.type", parameterToString(*r.idType, ""))
	}
	if r.idValue != nil {
		localVarQueryParams.Add("id.value", parameterToString(*r.idValue, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

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
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RuntimeError
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiClusterServiceRotateAuthRequest struct {
	ctx        _context.Context
	ApiService *ClusterServiceApiService
	idValue    string
}

func (r ApiClusterServiceRotateAuthRequest) Execute() (map[string]interface{}, *_nethttp.Response, error) {
	return r.ApiService.ClusterServiceRotateAuthExecute(r)
}

/*
ClusterServiceRotateAuth RotateAuth rotates the bearer token used for a cluster

	@param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param idValue value holds the cluster server URL or cluster name
	@return ApiClusterServiceRotateAuthRequest
*/
func (a *ClusterServiceApiService) ClusterServiceRotateAuth(ctx _context.Context, idValue string) ApiClusterServiceRotateAuthRequest {
	return ApiClusterServiceRotateAuthRequest{
		ApiService: a,
		ctx:        ctx,
		idValue:    idValue,
	}
}

// Execute executes the request
//
//	@return map[string]interface{}
func (a *ClusterServiceApiService) ClusterServiceRotateAuthExecute(r ApiClusterServiceRotateAuthRequest) (map[string]interface{}, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod  = _nethttp.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue map[string]interface{}
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ClusterServiceApiService.ClusterServiceRotateAuth")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/clusters/{id.value}/rotate-auth"
	localVarPath = strings.Replace(localVarPath, "{"+"id.value"+"}", _neturl.PathEscape(parameterToString(r.idValue, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

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
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RuntimeError
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiClusterServiceUpdateRequest struct {
	ctx           _context.Context
	ApiService    *ClusterServiceApiService
	idValue       string
	body          *V1alpha1Cluster
	updatedFields *[]string
	idType        *string
}

func (r ApiClusterServiceUpdateRequest) Body(body V1alpha1Cluster) ApiClusterServiceUpdateRequest {
	r.body = &body
	return r
}
func (r ApiClusterServiceUpdateRequest) UpdatedFields(updatedFields []string) ApiClusterServiceUpdateRequest {
	r.updatedFields = &updatedFields
	return r
}

// type is the type of the specified cluster identifier ( \&quot;server\&quot; - default, \&quot;name\&quot; ).
func (r ApiClusterServiceUpdateRequest) IdType(idType string) ApiClusterServiceUpdateRequest {
	r.idType = &idType
	return r
}

func (r ApiClusterServiceUpdateRequest) Execute() (V1alpha1Cluster, *_nethttp.Response, error) {
	return r.ApiService.ClusterServiceUpdateExecute(r)
}

/*
ClusterServiceUpdate Update updates a cluster

	@param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param idValue value holds the cluster server URL or cluster name
	@return ApiClusterServiceUpdateRequest
*/
func (a *ClusterServiceApiService) ClusterServiceUpdate(ctx _context.Context, idValue string) ApiClusterServiceUpdateRequest {
	return ApiClusterServiceUpdateRequest{
		ApiService: a,
		ctx:        ctx,
		idValue:    idValue,
	}
}

// Execute executes the request
//
//	@return V1alpha1Cluster
func (a *ClusterServiceApiService) ClusterServiceUpdateExecute(r ApiClusterServiceUpdateRequest) (V1alpha1Cluster, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod  = _nethttp.MethodPut
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue V1alpha1Cluster
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ClusterServiceApiService.ClusterServiceUpdate")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v1/clusters/{id.value}"
	localVarPath = strings.Replace(localVarPath, "{"+"id.value"+"}", _neturl.PathEscape(parameterToString(r.idValue, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}
	if r.body == nil {
		return localVarReturnValue, nil, reportError("body is required and must be specified")
	}

	if r.updatedFields != nil {
		t := *r.updatedFields
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("updatedFields", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("updatedFields", parameterToString(t, "multi"))
		}
	}
	if r.idType != nil {
		localVarQueryParams.Add("id.type", parameterToString(*r.idType, ""))
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
	// body params
	localVarPostBody = r.body
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RuntimeError
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
