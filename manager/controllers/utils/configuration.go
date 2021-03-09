// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/onsi/ginkgo"
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
	VaultModulesRole                    string = "VAULT_MODULES_ROLE"
)

// GetSystemNamespace returns the namespace of control plane
func GetSystemNamespace() string {
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "m4d-system"
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

// GetModulesRole returns the modules assigned authentification role for accessing dataset credentials
func GetModulesRole() string {
	return os.Getenv(VaultModulesRole)
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

// GetVaultAuthTTL returns the amount of time the authorization issued by vault is valid
func GetVaultAuthTTL() string {
	return os.Getenv(VaultTTLKey)
}

// GetVaultAuth returns the authentication method for vault connection
func GetVaultAuth() string {
	return os.Getenv(VaultAuthKey)
}

// GetCredentialsManagerServiceAddress returns the address where credentials manager is running
func GetCredentialsManagerServiceAddress() string {
	return os.Getenv(CredentialsManagerServiceAddressKey)
}

// GetDataCatalogServiceAddress returns the address where data catalog is running
func GetDataCatalogServiceAddress() string {
	return os.Getenv(CatalogConnectorServiceAddressKey)
}

func SetIfNotSet(key string, value string, t ginkgo.GinkgoTInterface) {
	if _, b := os.LookupEnv(key); !b {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Could not set environment variable %s", key)
		}
	}
}

func DefaultTestConfiguration(t ginkgo.GinkgoTInterface) {
	SetIfNotSet(CatalogConnectorServiceAddressKey, "localhost:50085", t)
	SetIfNotSet(CredentialsManagerServiceAddressKey, "localhost:50085", t)
	SetIfNotSet(VaultAddressKey, "http://127.0.0.1:8200/", t)
	SetIfNotSet(VaultDatasetMountKey, "v1/sys/mounts/m4d/dataset-creds", t)
	SetIfNotSet(VaultUserMountKey, "v1/sys/mounts/m4d/user-creds", t)
	SetIfNotSet(VaultDatasetHomeKey, "m4d/dataset-creds/", t)
	SetIfNotSet(VaultUserHomeKey, "m4d/user-creds/", t)
	SetIfNotSet("RUN_WITHOUT_VAULT", "1", t)
	SetIfNotSet("ENABLE_WEBHOOKS", "false", t)
	SetIfNotSet("CONNECTION_TIMEOUT", "120", t)
	SetIfNotSet("MAIN_POLICY_MANAGER_CONNECTOR_URL", "localhost:50090", t)
	SetIfNotSet("MAIN_POLICY_MANAGER_NAME", "MOCK", t)
	SetIfNotSet("USE_EXTENSIONPOLICY_MANAGER", "false", t)
	SetIfNotSet("SECRET_PROVIDER_URL", "http://secret-provider.m4d-system.svc.cluster.local:5555/get-secret", t)
	SetIfNotSet("SECRET_PROVIDER_ROLE", "demo", t)
}
