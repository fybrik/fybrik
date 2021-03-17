// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"net/url"
)

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
// It is of the form v1/m4d-system/dataset-creds/...
func GetSecretPath(assetID string) string {
	base := "v1/" + GetVaultDatasetHome()
	return fmt.Sprintf("%s%s", base, url.PathEscape(assetID))
}

// GetAuthPath returns the auth method path to use
// It is of the form v1/auth/<auth path>/login
func GetAuthPath(authPath string) string {
	fullAuthPath := fmt.Sprintf("v1/auth/%s/login", authPath)
	return fullAuthPath
}
