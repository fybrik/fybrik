// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"errors"
	"fmt"
	"strings"
)

// The path of the Vault plugin to use to retrieve dataset credentials stored in kubernetes secret.
// vault-plugin-secrets-kubernetes-reader plugin is used for this purpose and is enabled
// in kubernetes-secrets path. (https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader)
// TODO: pass the plugin path in fybrik-config ConfigMap
const vaultPluginPath = "kubernetes-secrets"
const secretPath = "/v1/" + vaultPluginPath + "/"

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
func PathForReadingKubeSecret(secretNamespace, secretName string) string {
	pluginPath := "/v1/" + vaultPluginPath + "/"
	// Construct the path to the secret in Vault that holds the dataset credentials
	secretPath := fmt.Sprintf("%s%s?namespace=%s", pluginPath, secretName, secretNamespace)
	return secretPath
}

// Given a path to Vault secret that holds dataset credentials
// return the name of the secret and its namespace
// for example, for vault secret path:
// "/v1/kubernetes-secrets/my-secret?namespace=default"
// the returned values will be my-secret and default
func GetKubeSecretDetailsFromVaultPath(credentialsPath string) (string, string, error) {
	str := strings.SplitAfter(credentialsPath, secretPath)
	if str[0] == credentialsPath {
		return "", "", errors.New("unexpected vault path format: wrong prefix " + credentialsPath)
	}

	parts := strings.Split(str[1], "?namespace=")
	if len(parts) != 2 {
		return "", "", errors.New("unexpected vault path format")
	}
	return parts[0], parts[1], nil
}
