// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package slices

func ContainsString(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}
