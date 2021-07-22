// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import "fmt"

// The path of the Vault plugin to use to retrieve dataset credentials stored in kubernetes secret.
// vault-plugin-secrets-kubernetes-reader plugin is used for this purpose and is enabled
// in kubernetes-secrets path. (https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader)
// TODO: pass the plugin path in fybrik-config ConfigMap
const vaultPluginPath = "kubernetes-secrets"

// PathForReadingKubeSecret returns the path to Vault secret that holds dataset credentials
// stored in kubernetes secret.
// Vault plugin vault-plugin-secrets-kubernetes-reader is used for reading kubernetes secret
// (https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader)
// The path contains the following parts:
// - pluginPath is the Vault path where vault-plugin-secrets-kubernetes-reader plugin is enabled.
// - secret name
// - secret namespace
// for example, for secret name my-secret and namespace default it will be of the form:
// "/v1/kubernetes-secrets/my-secret?namespace=default"
func PathForReadingKubeSecret(secretNamespace string, secretName string) string {
	pluginPath := "/v1/" + vaultPluginPath + "/"
	// Construct the path to the secret in Vault that holds the dataset credentials
	secretPath := fmt.Sprintf("%s%s?namespace=%s", pluginPath, secretName, secretNamespace)
	return secretPath
}
