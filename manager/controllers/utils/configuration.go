// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
)

// Attributes that are defined in a config map or the runtime environment
const (
	CatalogConnectorServiceAddressKey   string = "CATALOG_CONNECTOR_URL"
	CredentialsManagerServiceAddressKey string = "CREDENTIALS_CONNECTOR_URL"
	VaultAddressKey                     string = "VAULT_ADDRESS"
	VaultSecretKey                      string = "VAULT_TOKEN"
	VaultDatasetMountKey                string = "VAULT_DATASET_MOUNT"
	VaultUserMountKey                   string = "VAULT_USER_MOUNT"
	VaultUserHomeKey                    string = "VAULT_USER_HOME"
	VaultDatasetHomeKey                 string = "VAULT_DATASET_HOME"
	VaultTTLKey                         string = "VAULT_TTL"
	VaultAuthKey                        string = "VAULT_AUTH"
	SecretProviderURL                   string = "SECRET_PROVIDER_URL"
	SecretProviderRole                  string = "SECRET_PROVIDER_ROLE"
)

// GetSystemNamespace returns the namespace of control plane
func GetSystemNamespace() string {
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "default"
}

// GetSecretProviderURL returns the path to secret provider
// A credentials path should begin with this URL
func GetSecretProviderURL() string {
	return os.Getenv(SecretProviderURL)
}

// GetSecretProviderRole returns the assigned authentification role for accessing dataset credentials
func GetSecretProviderRole() string {
	return os.Getenv(SecretProviderRole)
}

// GetVaultAuthTTL returns the amount of time the authorization issued by vault is valid
func GetVaultAuthTTL() string {
	return os.Getenv(VaultTTLKey)
}

// GetVaultAuthService returns the authentication service that was chosen for use with vault,
// and the configuration options for it.
// Vault support multiple different types of authentication such as java web tokens (jwt), github, aws ...
func GetVaultAuthService() (string, api.EnableAuthOptions) {
	auth := os.Getenv(VaultAuthKey)
	options := api.EnableAuthOptions{
		Type: auth,
	}
	return auth, options
}

// GetVaultAddress returns the address and port of the vault system,
// which is used for managing data set credentials
func GetVaultAddress() string {
	return os.Getenv(VaultAddressKey)
}

// GetVaultUserHome returns the home directory in vault where the user credentials for external systems access by the m4d are stored
// All credentials will be in sub-directories of this directory in the form of system/compute_label
func GetVaultUserHome() string {
	return os.Getenv(VaultUserHomeKey)
}

// GetVaultDatasetHome returns the home directory in vault of where dataset credentials are stored.
// All credentials will be in sub-directories of this directory in the form of catalog_id/dataset_id
func GetVaultDatasetHome() string {
	return os.Getenv(VaultDatasetHomeKey)
}

// GetVaultUserMountPath returns the mount directory in vault of where user credentials for the external systems accessed by the m4d are stored.
func GetVaultUserMountPath() string {
	return os.Getenv(VaultUserMountKey)
}

// GetVaultDatasetMountPath returns the mount directory in vault of where dataset credentials are stored.
// All credentials will be in sub-directories of this directory in the form of catalog_id/dataset_id
func GetVaultDatasetMountPath() string {
	return os.Getenv(VaultDatasetMountKey)
}

// GetVaultToken returns the token this module uses to authenticate with vault
func GetVaultToken() string {
	return os.Getenv(VaultSecretKey)
}

// GetCredentialsManagerServiceAddress returns the address where credentials manager is running
func GetCredentialsManagerServiceAddress() string {
	return os.Getenv(CredentialsManagerServiceAddressKey)
}

// GetDataCatalogServiceAddress returns the address where data catalog is running
func GetDataCatalogServiceAddress() string {
	return os.Getenv(CatalogConnectorServiceAddressKey)
}
