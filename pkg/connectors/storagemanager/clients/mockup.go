// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"emperror.dev/errors"

	"fybrik.io/fybrik/pkg/model/storagemanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// mockup implementation of the storage manager client
type mockupStorageManager struct {
}

func NewMockupStorageManager() StorageManagerInterface {
	return &mockupStorageManager{}
}

func (m *mockupStorageManager) AllocateStorage(request *storagemanager.AllocateStorageRequest) (*storagemanager.AllocateStorageResponse,
	error) {
	if request == nil {
		return nil, errors.New("bad request")
	}
	resp := &storagemanager.AllocateStorageResponse{Connection: &taxonomy.Connection{Name: request.AccountType,
		AdditionalProperties: request.AccountProperties.Properties}}
	return resp, nil
}

func (m *mockupStorageManager) DeleteStorage(request *storagemanager.DeleteStorageRequest) error {
	return nil
}

func (m *mockupStorageManager) GetSupportedConnectionTypes() (*storagemanager.GetSupportedConnectionsResponse, error) {
	return &storagemanager.GetSupportedConnectionsResponse{ConnectionTypes: []taxonomy.ConnectionType{"mysql", "db2", "s3"}}, nil
}

func (m *mockupStorageManager) Close() error {
	return nil
}
