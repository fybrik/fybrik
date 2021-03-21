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
	catalogTaxName = "data_catalog_schema.json"

	geographyGood1 = "{\"geography\": {\"name\": \"Turkey\", \"geography_type\": \"country\"}}"
	geographyGood2 = "{\"geography\": {\"name\": \"Turkey\"}}"
	geographyBad1  = "{\"geography\": {\"name\": \"BlaBla\", \"geography_type\": \"country\"}}"

	dataSensitivityGood = "{\"data_sensitivity\":\"confidential\"}"
	dataSensitivityBad  = "{\"data_sensitivity\":\"whatever\"}"

	dataFormatGood = "{\"data_format\":\"arrow\"}"
	dataFormatBad  = "{\"data_format\":\"whatever\"}"

	dataTypeGood = "{\"data_type\":\"tabular\"}"
	dataTypeBad  = "{\"data_format\":\"whatever\"}"

	resourceGood1 = "{\"resource\": {\"name\":\"file1\", \"category\":\"tabular\", \"location\":{\"name\":\"Turkey\"}, \"tags\":[{\"a\":\"confidential\"}, \"confidential\"], \"columns\":[{\"name\":\"col1\"}, {\"name\":\"col2\",\"tags\":[{\"data_sensitivity\":\"confidential\"}]}]}}"
	resourceGood2 = "{\"resource\": {\"name\":\"file1\", \"category\":\"tabular\", \"location\":{\"name\":\"Turkey\"}, \"tags\":[{\"a\":\"confidential\"}, \"confidential\"], \"columns\":[{\"name\":\"col1\"}, {\"name\":\"col2\", \"tags\":[\"confidential\"]}]}}"
	resourceBad1  = "{\"resource\": {\"name\":\"file1\", \"category\":\"tabular\", \"location\":{\"name\":\"Turkey\"}, \"tags\":[{\"a\":\"whatever\"}]}}}"
	resourceBad2  = "{\"resource\": {\"name\":\"file1\", \"category\":\"tabular\", \"location\":{\"name\":\"Turkey\"}, \"tags\":[{\"a\":\"confidential\"}], \"columns\":[{\"name\":\"col1\"}, {\"name\":\"col2\"}, \"tags\":{\"data_sensitivity\":\"whatever\"}]}]}"
)

func validateJSON(t *testing.T, jsonData string, testName string, expectedValid bool) {
	path, err := filepath.Abs(catalogTaxName)
	assert.Nil(t, err)

	taxonomyLoader := gojsonschema.NewReferenceLoader("file://" + path)
	documentLoader := gojsonschema.NewStringLoader(jsonData)
	result, err := gojsonschema.Validate(taxonomyLoader, documentLoader)
	assert.Nil(t, err)

	fmt.Printf("%s valid document: %t\n", testName, result.Valid())

	if expectedValid {
		assert.True(t, result.Valid())
	} else {
		fmt.Printf("The document is not valid.  Discrepencies: \n")
		for _, disc := range result.Errors() {
			fmt.Printf("- %s\n", disc)
		}
		assert.False(t, result.Valid())
	}

}

func TestCatalogTaxonomy(t *testing.T) {
	validateJSON(t, geographyGood1, "geographyGood1", true)
	validateJSON(t, geographyGood2, "geographyGood2", true)
	validateJSON(t, geographyBad1, "geographyBad1", false)

	validateJSON(t, dataSensitivityGood, "dataSensitivityGood", true)
	validateJSON(t, dataSensitivityBad, "dataSensitivityBad", false)

	validateJSON(t, dataFormatGood, "dataFormatGood", true)
	validateJSON(t, dataFormatBad, "dataFormatBad", false)

	validateJSON(t, dataTypeGood, "dataTypeGood", true)
	validateJSON(t, dataTypeBad, "dataTypeBad", false)

	validateJSON(t, resourceGood1, "resourceGood1", true)
	validateJSON(t, resourceGood2, "resourceGood2", true)
	validateJSON(t, resourceBad1, "resourceBad1", false)
	validateJSON(t, resourceBad1, "resourceBad2", false)
}
