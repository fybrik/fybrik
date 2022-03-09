// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"

	openapiclient "fybrik.io/fybrik/pkg/connectors/datacatalog/openapiclient"
	"fybrik.io/fybrik/pkg/model/datacatalog"
)

var _ DataCatalog = (*openAPIDataCatalog)(nil)

type openAPIDataCatalog struct {
	name   string
	client *openapiclient.APIClient
}

// NewopenApiDataCatalog creates a DataCatalog facade that connects to a openApi service
func NewOpenAPIDataCatalog(name, connectionURL string, connectionTimeout time.Duration) DataCatalog {
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
	}
	apiClient := openapiclient.NewAPIClient(configuration)

	return &openAPIDataCatalog{
		name:   name,
		client: apiClient,
	}
}

func (m *openAPIDataCatalog) GetAssetInfo(in *datacatalog.GetAssetRequest, creds string) (*datacatalog.GetAssetResponse, error) {
	resp, _, err := m.client.DefaultApi.GetAssetInfo(context.Background()).XRequestDatacatalogCred(creds).GetAssetRequest(*in).Execute()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("get asset info from %s failed", m.name))
	}
	return resp, nil
}

func (m *openAPIDataCatalog) CreateAssetInfo(in *datacatalog.CreateAssetRequest, creds string) (*datacatalog.CreateAssetResponse, error) {
	resp, _, err := m.client.DefaultApi.CreateAssetInfo(context.Background()).
		XRequestDatacatalogWriteCred(creds).CreateAssetRequest(*in).Execute()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("create asset info from %s failed", m.name))
	}
	return resp, nil
}

func (m *openAPIDataCatalog) Close() error {
	return nil
}
