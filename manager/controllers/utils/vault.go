// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"net/url"
)

// The path of the Vault plugin to use to retrieve dataset credentials stored in kubernetes secret.
// vault-plugin-secrets-kubernetes-reader plugin is used for this purpose and is enabled
// in kubernetes-secrets path. (https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader)
// TODO: pass the plugin path in m4d-config ConfigMap
const vaultPluginPath = "kubernetes-secrets"

// GetFullCredentialsPath returns the path to be used for credentials retrieval
func GetFullCredentialsPath(secretName string) string {
	base := GetSecretProviderURL() + "?role=" + GetSecretProviderRole() + "&secret_name="
	return fmt.Sprintf("%s%s", base, url.QueryEscape(secretName))
}

// GetDatasetVaultPath returns the path that can be used externally to retrieve a dataset's credentials
// It is of the form <secret-provider-url>/v1/m4d-system/dataset-creds/...
func GetDatasetVaultPath(assetID string) string {
	secretName := "/v1/" + GetVaultDatasetHome() + assetID
	return GetFullCredentialsPath(secretName)
}

// GetSecretPath returns the path to the secret that holds the dataset's credentials
// It is of the form /v1/m4d-system/dataset-creds/...
func GetSecretPath(assetID string) string {
	base := "/v1/" + GetVaultDatasetHome()
	return fmt.Sprintf("%s%s", base, url.PathEscape(assetID))
}

// VaultPathForReadingKubeSecret returns the path to Vault secret that holds dataset credentials
// stored in kubernetes secret.
// Vault plugin vault-plugin-secrets-kubernetes-reader is used for reading the kubernetes secret and
// returning the dataset credentials.
// The path contains the following parts:
// - pluginPath is the Vault path where vault-plugin-secrets-kubernetes-reader plugin is enabled.
// - secret name
// - secret namespace
// for example, for secret name my-secret and namespace default it will be of the form:
// "/v1/kubernetes-secrets/my-secret?namespace=default"
func VaultPathForReadingKubeSecret(secretNamespace string, secretName string) string {
	pluginPath := "/v1/" + vaultPluginPath + "/"
	// Construct the path to the secret in Vault that holds the dataset credentials
	secretPath := fmt.Sprintf("%s%s?namespace=%s", pluginPath, secretName, secretNamespace)
	return secretPath
}

// GetAuthPath returns the auth method path to use
// It is of the form v1/auth/<auth path>/login
func GetAuthPath(authPath string) string {
	fullAuthPath := fmt.Sprintf("/v1/auth/%s/login", authPath)
	return fullAuthPath
}
