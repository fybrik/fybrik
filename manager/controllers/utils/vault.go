// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
)

// GetAuthPath returns the auth method path to use
// It is of the form v1/auth/<auth path>/login
// TODO - Different credentials for different data flows (read, write, delete)
func GetAuthPath(authPath string) string {
	fullAuthPath := fmt.Sprintf("/v1/auth/%s/login", authPath)
	return fullAuthPath
}
