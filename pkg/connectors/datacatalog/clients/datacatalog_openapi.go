// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"
	"os"
	"time"

	openapiclient "fybrik.io/fybrik/pkg/connectors/datacatalog/openapiclient"
	datacatalogTaxonomyModels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
)

var _ DataCatalog = (*openAPIDataCatalog)(nil)

type openAPIDataCatalog struct {
	name   string
	client *openapiclient.APIClient
}

// NewopenApiDataCatalog creates a DataCatalog facade that connects to a openApi service
func NewOpenAPIDataCatalog(name string, connectionURL string, connectionTimeout time.Duration) (DataCatalog, error) {
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
	}, nil
}

func (m *openAPIDataCatalog) GetAssetInfo(in *datacatalogTaxonomyModels.DataCatalogRequest, creds string) (*datacatalogTaxonomyModels.DataCatalogResponse, error) {
	resp, r, err := m.client.DefaultApi.GetAssetInfoPost(context.Background()).XRequestDataCatalogCred(creds).DataCatalogRequest(*in).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.GetAssetInfoPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAssetInfoPost`: DataCatalogResponse
	fmt.Fprintf(os.Stdout, "Response from `DefaultApi.GetAssetInfoPost` in GetAssetInfo of datacatalog_openapi.go: %v\n", resp)

	// resp, r, err := m.client.DefaultApi.GetPoliciesDecisionsPost(context.Background()).XRequestCred(creds).DataCatalogRequest(*in).Execute()
	// resp, r, err := m.client.DefaultApi.GetPoliciesDecisions(context.Background()).Input(*in).Creds(creds).Execute()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.GetPoliciesDecisions``: %v\n", err)
	// 	fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	// 	return nil, errors.Wrap(err, fmt.Sprintf("get policies decisions from %s failed", m.name))
	// }
	// response from `GetPoliciesDecisions`: []DataCatalogResponse
	// log.Println("1Response from `DefaultApi.GetPoliciesDecisions`: \n", resp)
	return &resp, nil
}

func (m *openAPIDataCatalog) Close() error {
	return nil
}
