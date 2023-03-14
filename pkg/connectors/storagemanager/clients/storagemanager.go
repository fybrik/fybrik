// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"io"

	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/model/storagemanager"
)

// StorageManagerInterface is an interface for storage management
type StorageManagerInterface interface {
	// AllocateStorage allocates storage based on the selected storage account by invoking the specific implementation agent
	// returns a Connection object in case of success, and an error - otherwise
	AllocateStorage(request *storagemanager.AllocateStorageRequest) (*storagemanager.AllocateStorageResponse, error)
	// DeleteStorage deletes the allocated storage
	DeleteStorage(request *storagemanager.DeleteStorageRequest) error
	// GetSupportedStorageTypes returns a list of supported connection types
	GetSupportedStorageTypes() (*storagemanager.GetSupportedStorageTypesResponse, error)
	io.Closer
}

func NewStorageManager() (StorageManagerInterface, error) {
	return NewOpenAPIStorageManager(environment.GetStorageManagerAddress()), nil
}
