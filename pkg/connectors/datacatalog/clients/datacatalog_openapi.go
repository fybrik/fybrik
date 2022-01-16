// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"time"

	openapiclient "fybrik.io/fybrik/pkg/connectors/datacatalog/openapiclient"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"github.com/rs/zerolog"
)

var _ DataCatalog = (*openAPIDataCatalog)(nil)

type openAPIDataCatalog struct {
	name   string
	client *openapiclient.APIClient
	log    zerolog.Logger
}

// NewopenApiDataCatalog creates a DataCatalog facade that connects to a openApi service
func NewOpenAPIDataCatalog(name string, connectionURL string, connectionTimeout time.Duration, log zerolog.Logger) DataCatalog {
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
		log:    log,
	}
}

func (m *openAPIDataCatalog) GetAssetInfo(in *datacatalog.GetAssetRequest, creds string) (*datacatalog.GetAssetResponse, error) {
	log := m.log.With().Str(logging.DATASETID, string(in.AssetID)).Logger()
	resp, r, err := m.client.DefaultApi.GetAssetInfoPost(context.Background()).XRequestDataCatalogCred(creds).DataCatalogRequest(*in).Execute()
	if err != nil {
		log.Error().Err(err).Msg("error when calling `DefaultApi.GetAssetInfoPost`")
		logging.LogStructure("HTTP response", r, log, false, false)
	}
	// response from `GetAssetInfoPost`: DataCatalogResponse
	logging.LogStructure("datacatalog_openapi response", resp, log, false, false)
	return &resp, nil
}

func (m *openAPIDataCatalog) Close() error {
	return nil
}
