// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"emperror.dev/errors"

	"fybrik.io/fybrik/pkg/connectors/datacatalog/openapiclient"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/tls"
)

// ErrorMessages that are reported to the user
const (
	AssetIDNotFound       string = "the asset does not exist"
	AccessForbidden       string = "no permissions to access the data"
	DataStoreNotSupported string = "the asset data store is not supported"
)

var _ DataCatalog = (*openAPIDataCatalog)(nil)

type openAPIDataCatalog struct {
	name   string
	client *openapiclient.APIClient
}

// NewopenApiDataCatalog creates a DataCatalog facade that connects to a openApi service

func NewOpenAPIDataCatalog(name, connectionURL string) DataCatalog {
	log := logging.LogInit(logging.SETUP, "datacatalog client")
	configuration := &openapiclient.Configuration{
		DefaultHeader: make(map[string]string),
		UserAgent:     "OpenAPI-Generator/1.0.0/go",
		Debug:         false,
		Servers: openapiclient.ServerConfigurations{
			{
				URL:         connectionURL,
				Description: "No description provided",
			},
		},
		OperationServers: map[string]openapiclient.ServerConfigurations{},
		HTTPClient:       tls.GetHTTPClient(&log).StandardClient(),
	}
	apiClient := openapiclient.NewAPIClient(configuration)

	return &openAPIDataCatalog{
		name:   name,
		client: apiClient,
	}
}

// getDetailedError generates an error from the response body JSON if available,
// otherwise it extend the base error (e.g., 400 Bad Request) with the given message string
func getDetailedError(httpResponse *http.Response, baseError error, defaultMsg string) error {
	if bodyBytes, errRead := io.ReadAll(httpResponse.Body); errRead == nil && len(bodyBytes) > 0 {
		return errors.New(string(bodyBytes))
	}
	return errors.Wrap(baseError, defaultMsg)
}

func (m *openAPIDataCatalog) GetAssetInfo(in *datacatalog.GetAssetRequest, creds string) (*datacatalog.GetAssetResponse, error) {
	printErr := func() string { return fmt.Sprintf("get asset info from %s failed", m.name) }
	resp, httpResponse, err :=
		m.client.DefaultApi.GetAssetInfo(context.Background()).XRequestDatacatalogCred(creds).GetAssetRequest(*in).Execute()

	if httpResponse == nil {
		if err != nil {
			return nil, errors.Wrap(err, printErr())
		}
		return nil, errors.New(printErr())
	}
	defer httpResponse.Body.Close()
	// special errors equivalent to Deny response
	if httpResponse.StatusCode == http.StatusForbidden {
		return nil, errors.New(AccessForbidden)
	}
	if httpResponse.StatusCode == http.StatusNotFound {
		return nil, errors.New(AssetIDNotFound)
	}
	if err != nil {
		return nil, getDetailedError(httpResponse, err, printErr())
	}
	return &resp, nil
}

func (m *openAPIDataCatalog) CreateAsset(in *datacatalog.CreateAssetRequest, creds string) (*datacatalog.CreateAssetResponse, error) {
	printErr := func() string { return fmt.Sprintf("create asset info from %s failed", m.name) }
	resp, httpResponse, err := m.client.DefaultApi.CreateAsset(context.Background()).
		XRequestDatacatalogWriteCred(creds).CreateAssetRequest(*in).Execute()
	if httpResponse == nil {
		if err != nil {
			return nil, errors.Wrap(err, printErr())
		}
		return nil, errors.New(printErr())
	}
	defer httpResponse.Body.Close()

	if err != nil {
		return nil, getDetailedError(httpResponse, err, printErr())
	}
	return &resp, nil
}

//nolint:dupl
func (m *openAPIDataCatalog) DeleteAsset(in *datacatalog.DeleteAssetRequest, creds string) (*datacatalog.DeleteAssetResponse, error) {
	printErr := func() string { return fmt.Sprintf("delete asset info from %s failed", m.name) }
	resp, httpResponse, err :=
		m.client.DefaultApi.DeleteAsset(context.Background()).XRequestDatacatalogCred(creds).DeleteAssetRequest(*in).Execute()
	if httpResponse == nil {
		if err != nil {
			return nil, errors.Wrap(err, printErr())
		}
		return nil, errors.New(printErr())
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusNotFound {
		return nil, errors.New(AssetIDNotFound)
	}
	if err != nil {
		return nil, getDetailedError(httpResponse, err, printErr())
	}
	return &resp, nil
}

//nolint:dupl
func (m *openAPIDataCatalog) UpdateAsset(in *datacatalog.UpdateAssetRequest, creds string) (*datacatalog.UpdateAssetResponse, error) {
	resp, httpResponse, err := m.client.DefaultApi.UpdateAsset(
		context.Background()).XRequestDatacatalogUpdateCred(creds).UpdateAssetRequest(*in).Execute()
	printErr := func() string { return fmt.Sprintf("update asset info from %s failed", m.name) }
	if httpResponse == nil {
		if err != nil {
			return nil, errors.Wrap(err, printErr())
		}
		return nil, errors.New(printErr())
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusNotFound {
		return nil, errors.New(AssetIDNotFound)
	}
	if err != nil {
		return nil, getDetailedError(httpResponse, err, printErr())
	}
	return &resp, nil
}

func (m *openAPIDataCatalog) Close() error {
	return nil
}
