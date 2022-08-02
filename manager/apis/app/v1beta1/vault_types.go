// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta1

// Holds details for retrieving credentials from Vault store.
type Vault struct {
	// Role is the Vault role used for retrieving the credentials
	// +required
	Role string `json:"role"`
	// SecretPath is the path of the secret holding the Credentials in Vault
	// +required
	SecretPath string `json:"secretPath"`
	// Address is Vault address
	// +required
	Address string `json:"address"`
	// AuthPath is the path to auth method i.e. kubernetes
	// +required
	AuthPath string `json:"authPath"`
}
