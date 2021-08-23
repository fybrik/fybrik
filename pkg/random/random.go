// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package random

import (
	"crypto/rand"
	"encoding/hex"
)

// ref: https://sosedoff.com/2014/12/15/generate-random-hex-string-in-go.html
func Hex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
