// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
)

// GetDataDir returns the directory where the data resides.
func GetDataDir() string {
	return os.Getenv("DATA")
}
