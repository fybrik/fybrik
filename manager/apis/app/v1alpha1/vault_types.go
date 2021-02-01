// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// Holds details for retrieving credentials from Vault store.
type Vault struct {
	// Role is the Vault role used to retrieving credentials
	// +required
	Role string `json:"role"`
	// Path is the credentials path in Vault
	// +required
	Path string `json:"path"`
	// Address is Vault address
	// +required
	Address string `json:"address"`
	// AuthPath is the authentication method used to acees Vault i.e. kubernetes
	// +required
	AuthPath string `json:"authPath"`
}
