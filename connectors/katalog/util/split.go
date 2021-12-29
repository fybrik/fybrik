// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"fmt"
	"strings"
)

func SplitNamespacedName(value string) (namespace string, name string, err error) {
	identifier := strings.SplitN(value, "/", 2)
	if len(identifier) != 2 {
		err = fmt.Errorf("expected <namespace>/<name> format but got %s", value)
		return
	}
	namespace, name = identifier[0], identifier[1]
	return
}
