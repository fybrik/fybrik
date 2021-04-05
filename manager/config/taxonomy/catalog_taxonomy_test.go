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

var (
	catalogTaxStructsName = "catalog.structs.schema.json"
	catalogTaxValsName    = "catalog.values.schema.json"

	geographyGood1 = "{\"geography\": {\"name\": \"Turkey\", \"geography_type\": \"country\"}}"
	geographyGood2 = "{\"geography\": {\"name\": \"Turkey\"}}"
	geographyBad1  = "{\"geography\": {\"name\": \"BlaBla\", \"geography_type\": \"country\"}}"

	dataFormatGood = "{\"data_format\":\"arrow\"}"
	dataFormatBad  = "{\"data_format\":\"whatever\"}"

	dataTypeGood = "{\"data_type\":\"tabular\"}"
	dataTypeBad  = "{\"data_format\":\"whatever\"}"

	// {"resource":{"name":"file1"}}
	resourceGoodNameOnly = "{\"resource\": {\"name\":\"file1\"}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}]}}
	resourceGoodNameGeo = "{\"resource\": {\"name\":\"file1\", \"tags\":[{\"geography\":{\"name\":\"Turkey\"}}]}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}], "columns":[{"name":"col1"}, {"name":"col2"}]}}
	resourceGoodCols = "{\"resource\":{\"name\":\"file1\", \"tags\":[{\"geography\":{\"name\":\"Turkey\"}}], \"columns\":[{\"name\":\"col1\"}, {\"name\":\"col2\"}]}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}], "columns":[{"name":"col1", "tags":[{"geography_name":"Turkey"}]}, {"name":"col2"}]}}
	resourceGoodColsTags = "{\"resource\":{\"name\":\"file1\", \"tags\":[{\"geography\":{\"name\":\"Turkey\"}}], \"columns\":[{\"name\":\"col1\", \"tags\":[{\"geography_name\":\"Turkey\"}]}, {\"name\":\"col2\"}]}}"

	// {"resource":{"tags":[{"geography":{"name":"Turkey"}}]}}
	resourceBadNoName = "{\"resource\": {\"tags\":[{\"geography\":{\"name\":\"Turkey\"}}]}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"xxx"}}]}}
	resourceBadInvalidGeo = "{\"resource\":{\"name\":\"file1\", \"tags\":[{\"geography\":{\"name\":\"xxx\"}}]}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}], "columns":[{"name":"col1", "tags":[{"geography_name":"xxx"}]}, {"name":"col2"}]}}
	resourceBadInvalidTagVal = "{\"resource\":{\"name\":\"file1\", \"tags\":[{\"geography\":{\"name\":\"Turkey\"}}], \"columns\":[{\"name\":\"col1\", \"tags\":[{\"geography_name\":\"xxx\"}]}, {\"name\":\"col2\"}]}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}], "columns":[{"name":"col1", "tags":[{"badkey":"Turkey"}]}, {"name":"col2"}]}}
	resourceGoodInvalidTagKey = "{\"resource\":{\"name\":\"file1\", \"tags\":[{\"geography\":{\"name\":\"Turkey\"}}], \"columns\":[{\"name\":\"col1\", \"tags\":[{\"badkey\":\"Turkey\"}]}, {\"name\":\"col2\"}]}}"
)

// validateTaxonomy loads a json schema taxonomy from the indicated file, and validates the jsonData against the taxonomy.
func validateTaxonomy(t *testing.T, taxonomyFile string, jsonData string, testName string, expectedValid bool) {
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

func TestCatalogTaxonomy(t *testing.T) {
	validateTaxonomy(t, catalogTaxValsName, dataFormatGood, "dataFormatGood", true)
	validateTaxonomy(t, catalogTaxValsName, dataFormatBad, "dataFormatBad", false)

	validateTaxonomy(t, catalogTaxValsName, dataTypeGood, "dataTypeGood", true)
	validateTaxonomy(t, catalogTaxValsName, dataTypeBad, "dataTypeBad", false)

	validateTaxonomy(t, catalogTaxStructsName, geographyGood1, "geographyGood1", true)
	validateTaxonomy(t, catalogTaxStructsName, geographyGood2, "geographyGood2", true)
	validateTaxonomy(t, catalogTaxStructsName, geographyBad1, "geographyBad1", false)

	validateTaxonomy(t, catalogTaxStructsName, resourceGoodNameOnly, "resourceGoodNameOnly", true)
	validateTaxonomy(t, catalogTaxStructsName, resourceGoodNameGeo, "resourceGoodNameGeo", true)
	validateTaxonomy(t, catalogTaxStructsName, resourceGoodCols, "resourceGoodCols", true)
	validateTaxonomy(t, catalogTaxStructsName, resourceGoodColsTags, "resourceGoodColsTags", true)

	validateTaxonomy(t, catalogTaxStructsName, resourceBadNoName, "resourceBadNoName", false)
	validateTaxonomy(t, catalogTaxStructsName, resourceBadInvalidGeo, "resourceBadInvalidGeo", false)
	validateTaxonomy(t, catalogTaxStructsName, resourceBadInvalidTagVal, "resourceBadInvalidTagVal", false)
	validateTaxonomy(t, catalogTaxStructsName, resourceGoodInvalidTagKey, "resourceBadInvalidTagKey", true)
}
