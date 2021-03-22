// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func EmptyIfNil(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

func EmptyArrayIfNil(val *[]string) []string {
	if val == nil {
		return []string{}
	}
	return *val
}

func DecodeToStruct(m interface{}, s interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	err = dec.Decode(s)
	if err != nil {
		return err
	}
	return nil
}

func SplitNamespacedName(value string) (namespace string, name string, err error) {
	identifier := strings.SplitN(value, "/", 2)
	if len(identifier) != 2 {
		err = fmt.Errorf("Expected <namespace>/<name> format but got %s", value)
		return
	}
	namespace, name = identifier[0], identifier[1]
	return
}
