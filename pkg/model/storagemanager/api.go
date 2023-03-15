// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package storagemanager

import (
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

type AllocateStorageRequest struct {
	// Type of the storage account, e.g., s3
	AccountType taxonomy.ConnectionType `json:"accountType"`
	// Account properties, e.g., endpoint
	AccountProperties taxonomy.StorageAccountProperties `json:"accountProperties"`
	// Reference to the secret with credentials
	Secret taxonomy.SecretRef `json:"secret,omitempty"`
	// Configuration options
	Opts Options `json:"options"`
}

type AllocateStorageResponse struct {
	// Connection object for the allocated storage
	Connection *taxonomy.Connection `json:"connection,omitempty"`
}

type DeleteStorageRequest struct {
	// Connection object representing storage to free
	Connection taxonomy.Connection `json:"connection"`
	// Reference to the secret with credentials
	Secret taxonomy.SecretRef `json:"secret,omitempty"`
	// Configuration options
	Opts Options `json:"options"`
}

type GetSupportedStorageTypesResponse struct {
	// connection types supported by StorageManager for storage allocation/deletion
	ConnectionTypes []taxonomy.ConnectionType `json:"connectionTypes"`
}
