// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

// ValidateTaxonomy loads a json schema taxonomy from the indicated file, and validates the jsonData against the taxonomy.
func ValidateTaxonomy(t *testing.T, taxonomyFile string, jsonData string, testName string, expectedValid bool) {
	path, err := filepath.Abs(taxonomyFile)
	assert.Nil(t, err)

	taxonomyLoader := gojsonschema.NewReferenceLoader("file://" + path)
	documentLoader := gojsonschema.NewStringLoader(jsonData)
	result, err := gojsonschema.Validate(taxonomyLoader, documentLoader)
	assert.Nil(t, err)

	if expectedValid {
		assert.True(t, result.Valid())
	} else {
		assert.False(t, result.Valid())
	}

	if (result.Valid() && !expectedValid) || (!result.Valid() && expectedValid) {
		fmt.Printf("%s unexpected result.  Taxonomy file %s.  Discrepencies: \n", testName, taxonomyFile)
		for _, disc := range result.Errors() {
			fmt.Printf("- %s\n", disc)
		}
		fmt.Printf("\n")
	}
}
