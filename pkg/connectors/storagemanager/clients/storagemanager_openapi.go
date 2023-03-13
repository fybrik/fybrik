// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"net/http"

	"emperror.dev/errors"
	"github.com/rs/zerolog"

	openapiclient "fybrik.io/fybrik/pkg/connectors/storagemanager/openapiclient"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/storagemanager"
	"fybrik.io/fybrik/pkg/tls"
)

// ErrorMessages that are reported to the user
const (
	StorageTypeNotSupported          string = "the requested storage type is not supported"
	StorageManagerCommunicationError string = "could not communicate with storage manager"
)

var _ StorageManagerInterface = (*openAPIStorageManager)(nil)

type openAPIStorageManager struct {
	Log    zerolog.Logger
	Client *openapiclient.APIClient
}

// NewOpenAPIStorageManager creates a StorageManagerInterface facade that connects to a openApi service
func NewOpenAPIStorageManager(address string) StorageManagerInterface {
	log := logging.LogInit(logging.SETUP, "storage manager client")
	configuration := &openapiclient.Configuration{
		DefaultHeader: make(map[string]string),
		UserAgent:     "OpenAPI-Generator/1.0.0/go",
		Debug:         false,
		Servers: openapiclient.ServerConfigurations{
			{
				URL:         address,
				Description: "No description provided",
			},
		},
		OperationServers: map[string]openapiclient.ServerConfigurations{},
		HTTPClient:       tls.GetHTTPClient(&log).StandardClient(),
	}
	apiClient := openapiclient.NewAPIClient(configuration)

	return &openAPIStorageManager{
		Log:    log,
		Client: apiClient,
	}
}

// storage allocation request
func (m *openAPIStorageManager) AllocateStorage(request *storagemanager.AllocateStorageRequest) (*storagemanager.AllocateStorageResponse,
	error) {
	resp, httpResponse, err :=
		m.Client.DefaultApi.AllocateStorage(context.Background()).AllocateStorageRequest(*request).Execute()
	if httpResponse == nil {
		if err != nil {
			return nil, err
		}
		return nil, errors.New(StorageManagerCommunicationError)
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusNotImplemented {
		return nil, errors.New(StorageTypeNotSupported)
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// storage deletion request
func (m *openAPIStorageManager) DeleteStorage(request *storagemanager.DeleteStorageRequest) error {
	httpResponse, err := m.Client.DefaultApi.DeleteStorage(context.Background()).DeleteStorageRequest(*request).Execute()
	if httpResponse == nil {
		if err != nil {
			return err
		}
		return errors.New(StorageManagerCommunicationError)
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusNotImplemented {
		return errors.New(StorageTypeNotSupported)
	}
	return err
}

// request to get supported connections
func (m *openAPIStorageManager) GetSupportedStorageTypes() (*storagemanager.GetSupportedStorageTypesResponse, error) {
	resp, httpResponse, err :=
		m.Client.DefaultApi.GetSupportedStorageTypes(context.Background()).Execute()
	if httpResponse == nil {
		if err != nil {
			return nil, err
		}
		return nil, errors.New(StorageManagerCommunicationError)
	}
	defer httpResponse.Body.Close()
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *openAPIStorageManager) Close() error {
	return nil
}
