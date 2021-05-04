// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"net/url"
)

// GetSecretPath returns the path to the secret that holds the dataset's credentials
// It is of the form /v1/m4d-system/dataset-creds/...
func GetSecretPath(assetID string) string {
	base := "/v1/" + GetVaultDatasetHome()
	return fmt.Sprintf("%s%s", base, url.PathEscape(assetID))
}

// GetAuthPath returns the auth method path to use
// It is of the form v1/auth/<auth path>/login
func GetAuthPath(authPath string) string {
	fullAuthPath := fmt.Sprintf("/v1/auth/%s/login", authPath)
	return fullAuthPath
}
