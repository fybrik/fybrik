// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DataStore contains the details for accesing the data that are sent by catalog connectors
// Credentials for accesing the data are stored in Vault, in the location represented by CredentialLocation property.
type DataStore struct {
	// CredentialLocation is used to obtain
	// the credentials from the credential management system - ex: vault
	// +optional
	CredentialLocation string `json:"credentialLocation,omitempty"`
	// Holds details for retrieving credentials by the modules from Vault store.
	// +optional
	Vault *Vault `json:"vault,omitempty"`
	// Connection has the relevant details for accesing the data (url, table, ssl, etc.)
	// +required
	Connection runtime.RawExtension `json:"connection"`
	// Format represents data format (e.g. parquet) as received from catalog connectors
	// +required
	Format string `json:"format"`
}
