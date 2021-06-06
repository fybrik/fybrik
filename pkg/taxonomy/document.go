// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
)

// ValidateDocument validates input json data against a taxonomy file
func ValidateDocument(taxonomyFile string, jsonData string) (*gojsonschema.Result, error) {
	path, err := filepath.Abs(taxonomyFile)
	if err != nil {
		return nil, err
	}

	taxonomyLoader := gojsonschema.NewReferenceLoader("file://" + path)
	documentLoader := gojsonschema.NewStringLoader(jsonData)
	result, err := gojsonschema.Validate(taxonomyLoader, documentLoader)
	if err != nil {
		return nil, err
	}

	return result, nil
}
