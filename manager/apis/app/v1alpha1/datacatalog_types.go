// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "github.com/mesh-for-data/mesh-for-data/pkg/serde"

// DataStore contains the details for accesing the data that are sent by catalog connectors
// Credentials for accesing the data are stored in Vault, in the location represented by Vault property.
type DataStore struct {
	// Holds details for retrieving credentials by the modules from Vault store.
	Vault Vault `json:"vault"`
	// Connection has the relevant details for accesing the data (url, table, ssl, etc.)
	// +required
	Connection serde.Arbitrary `json:"connection"`
	// Format represents data format (e.g. parquet) as received from catalog connectors
	// +required
	Format string `json:"format"`
}
