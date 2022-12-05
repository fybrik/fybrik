// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"

	"emperror.dev/errors"

	openapiclient "fybrik.io/fybrik/pkg/connectors/policymanager/openapiclient"
	"fybrik.io/fybrik/pkg/connectors/utils"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/policymanager"
)

var _ PolicyManager = (*openAPIPolicyManager)(nil)

type openAPIPolicyManager struct {
	name   string
	client *openapiclient.APIClient
}

// NewopenApiPolicyManager creates a PolicyManager facade that connects to a openApi service

func NewOpenAPIPolicyManager(name, connectionURL string) (PolicyManager, error) {
	log := logging.LogInit(logging.SETUP, "policymanager client")
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
		HTTPClient:       utils.GetHTTPClient(&log).StandardClient(),
	}
	apiClient := openapiclient.NewAPIClient(configuration)

	return &openAPIPolicyManager{
		name:   name,
		client: apiClient,
	}, nil
}

func (m *openAPIPolicyManager) GetPoliciesDecisions(in *policymanager.GetPolicyDecisionsRequest,
	creds string) (*policymanager.GetPolicyDecisionsResponse, error) {
	resp, _, err := m.client.DefaultApi.GetPoliciesDecisions(context.Background()).XRequestCred(creds).GetPolicyDecisionsRequest(*in).Execute()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("get policies decisions from %s failed", m.name))
	}
	return &resp, nil
}

func (m *openAPIPolicyManager) Close() error {
	return nil
}
