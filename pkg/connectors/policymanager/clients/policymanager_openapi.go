// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"emperror.dev/errors"

	"fybrik.io/fybrik/pkg/connectors/policymanager/openapiclient"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/policymanager"
	"fybrik.io/fybrik/pkg/tls"
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
		HTTPClient:       tls.GetHTTPClient(&log).StandardClient(),
	}
	apiClient := openapiclient.NewAPIClient(configuration)

	return &openAPIPolicyManager{
		name:   name,
		client: apiClient,
	}, nil
}

// getDetailedError generates an error from the response body JSON if available,
// otherwise it extend the base error (e.g., 400 Bad Request) with the given message string
func getDetailedError(httpResponse *http.Response, baseError error, defaultMsg string) error {
	if bodyBytes, errRead := io.ReadAll(httpResponse.Body); errRead == nil && len(bodyBytes) > 0 {
		return errors.New(string(bodyBytes))
	}
	return errors.Wrap(baseError, defaultMsg)
}

func (m *openAPIPolicyManager) GetPoliciesDecisions(in *policymanager.GetPolicyDecisionsRequest,
	creds string) (*policymanager.GetPolicyDecisionsResponse, error) {
	printErr := func() string { return fmt.Sprintf("get policies decisions from %s failed", m.name) }
	resp, httpResponse, err := m.client.DefaultApi.GetPoliciesDecisions(context.Background()).XRequestCred(creds).
		GetPolicyDecisionsRequest(*in).Execute()

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

func (m *openAPIPolicyManager) Close() error {
	return nil
}
