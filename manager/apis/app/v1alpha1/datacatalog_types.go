// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "fybrik.io/fybrik/pkg/model/taxonomy"

type DataFlow string

const (
	// ReadFlow indicates a data set is being read
	ReadFlow DataFlow = "read"

	// WriteFlow indicates a data set is being written
	WriteFlow DataFlow = "write"

	// DeleteFlow indicates a data set is being deleted
	DeleteFlow DataFlow = "delete"

	// CopyFlow indicates a data set is being copied
	CopyFlow DataFlow = "copy"
)

// DataStore contains the details for accesing the data that are sent by catalog connectors
// Credentials for accesing the data are stored in Vault, in the location represented by Vault property.
type DataStore struct {
	// Holds details for retrieving credentials by the modules from Vault store.
	// It is a map so that different credentials can be stored for the different DataFlow operations.
	Vault map[string]Vault `json:"vault"`
	// Connection has the relevant details for accesing the data (url, table, ssl, etc.)
	// +required
	Connection taxonomy.Connection `json:"connection"`
	// Format represents data format (e.g. parquet) as received from catalog connectors
	// +required
	Format string `json:"format"`
}
