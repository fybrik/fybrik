// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
	openapiclient "fybrik.io/fybrik/pkg/connectors/policymanager/openapiclient"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/policymanager"
	"github.com/rs/zerolog"
)

var _ PolicyManager = (*openAPIPolicyManager)(nil)

type openAPIPolicyManager struct {
	name   string
	client *openapiclient.APIClient
	log    zerolog.Logger
}

// NewopenApiPolicyManager creates a PolicyManager facade that connects to a openApi service
func NewOpenAPIPolicyManager(name string, connectionURL string, connectionTimeout time.Duration, log zerolog.Logger) (PolicyManager, error) {
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
		log:    log,
	}, nil
}

func (m *openAPIPolicyManager) GetPoliciesDecisions(in *policymanager.GetPolicyDecisionsRequest, creds string) (*policymanager.GetPolicyDecisionsResponse, error) {
	log := m.log.With().Str(logging.DATASETID, string(in.Resource.ID)).Logger()
	resp, r, err := m.client.DefaultApi.GetPoliciesDecisionsPost(context.Background()).XRequestCred(creds).PolicyManagerRequest(*in).Execute()
	// resp, r, err := m.client.DefaultApi.GetPoliciesDecisions(context.Background()).Input(*in).Creds(creds).Execute()
	if err != nil {
		log.Error().Err(err).Msg("error when calling `DefaultApi.GetPoliciesDecisions`")
		logging.LogStructure("HTTP response", r, log, false, false)
		return nil, errors.Wrap(err, fmt.Sprintf("get policies decisions from %s failed", m.name))
	}
	// response from `GetPoliciesDecisions`: []PolicymanagerResponse
	logging.LogStructure("policymanager_openapi response", resp, log, false, false)
	return &resp, nil
}

func (m *openAPIPolicyManager) Close() error {
	return nil
}
