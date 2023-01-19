// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import "fybrik.io/fybrik/pkg/model/taxonomy"

// DataStore contains the details for accessing the data that are sent by catalog connectors
// Credentials for accessing the data are stored in Vault, in the location represented by Vault property.
type DataStore struct {
	// Holds details for retrieving credentials by the modules from Vault store.
	// It is a map so that different credentials can be stored for the different DataFlow operations.
	// +optional
	Vault map[string]Vault `json:"vault,omitempty"`
	// Connection has the relevant details for accessing the data (url, table, ssl, etc.)
	// +required
	Connection taxonomy.Connection `json:"connection"`
	// Format represents data format (e.g. parquet) as received from catalog connectors
	// +optional
	Format taxonomy.DataFormat `json:"format,omitempty"`
}
