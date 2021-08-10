// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"emperror.dev/errors"
	openapiclient "fybrik.io/fybrik/pkg/connectors/openapiclient"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
)

var _ PolicyManager = (*openAPIPolicyManager)(nil)

type openAPIPolicyManager struct {
	name   string
	client *openapiclient.APIClient
}

// NewopenApiPolicyManager creates a PolicyManager facade that connects to a openApi service
func NewOpenAPIPolicyManager(name string, connectionURL string, connectionTimeout time.Duration) (PolicyManager, error) {
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

	return &openAPIPolicyManager{
		name:   name,
		client: apiClient,
	}, nil
}

func (m *openAPIPolicyManager) GetPoliciesDecisions(in *openapiclientmodels.PolicyManagerRequest, creds string) (*openapiclientmodels.PolicyManagerResponse, error) {
	resp, r, err := m.client.DefaultApi.GetPoliciesDecisionsPost(context.Background()).XRequestCred(creds).PolicyManagerRequest(*in).Execute()
	// resp, r, err := m.client.DefaultApi.GetPoliciesDecisions(context.Background()).Input(*in).Creds(creds).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.GetPoliciesDecisions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, errors.Wrap(err, fmt.Sprintf("get policies decisions from %s failed", m.name))
	}
	// response from `GetPoliciesDecisions`: []PolicymanagerResponse
	log.Println("1Response from `DefaultApi.GetPoliciesDecisions`: \n", resp)
	return &resp, nil
}

func (m *openAPIPolicyManager) Close() error {
	return nil
}
