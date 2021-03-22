// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package utils

import (
	"fmt"
)

// The path of the Vault plugin to use to retrieve the dataset credential.
// vault-plugin-secrets-kubernetes-reader plugin is used for this purpose and is enabled
// in kubernetes-secrets path. (https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader)
const vaultPluginPath = "kubernetes-secrets"

// VaultSecretPath returns the path to Vault secret that holds the dataset credential.
// The path contains the following parts:
// - pluginPath is the Vault path where vault-plugin-secrets-kubernetes-reader plugin is enabled.
// - secret name
// - secret namespace
// for example, for secret name my-secret and namespace default it will be of the form:
// "/v1/kubernetes-secrets/my-secret?namespace=default"
func VaultSecretPath(secretNamespace string, secretName string) string {
	pluginPath := "/v1/" + vaultPluginPath + "/"
	// Construct the path to the secret in Vault that holds the dataset credentials
	secretPath := fmt.Sprintf("%s%s?namespace=%s", pluginPath, secretName, secretNamespace)
	return secretPath
}
