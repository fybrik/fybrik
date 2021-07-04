// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"
)

var (
	catalogTaxStructsName = "../../charts/m4d/files/taxonomy/catalog.structs.schema.json"
	catalogTaxValsName    = "../../charts/m4d/files/taxonomy/catalog.values.schema.json"

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
	resourceGoodNameGeo = "{\"resource\": {\"name\":\"file1\", \"tags\":{\"residency\":\"Turkey\", \"asset\":\"PII\"}}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}], "columns":[{"name":"col1"}, {"name":"col2"}]}}
	resourceGoodCols = "{\"resource\":{\"name\":\"file1\", \"tags\":{\"residency\":\"Turkey\", \"asset\":\"PII\"}, \"columns\":[{\"name\":\"col1\"}, {\"name\":\"col2\"}]}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}], "columns":[{"name":"col1", "tags":[{"geography_name":"Turkey"}]}, {"name":"col2"}]}}
	resourceGoodColsTags = "{\"resource\":{\"name\":\"file1\", \"tags\":{\"residency\":\"Turkey\", \"asset\":\"PII\"}, \"columns\":[{\"name\":\"col1\", \"tags\":{\"residency\":\"Turkey\"}}, {\"name\":\"col2\"}]}}"

	// {"resource":{"tags":[{"geography":{"name":"Turkey"}}]}}
	resourceBadNoName = "{\"resource\": {\"tags\":{\"residency\":\"Turkey\", \"asset\":\"PII\"}}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"xxx"}}]}}
	// resourceBadInvalidGeo = "{\"resource\":{\"name\":\"file1\", \"tags\":{\"residency\": 12, \"asset\":\"PII\"}}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}], "columns":[{"name":"col1", "tags":[{"geography_name":"xxx"}]}, {"name":"col2"}]}}
	// resourceBadInvalidTagVal = "{\"resource\":{\"name\":\"file1\", \"tags\":[\"residency\":\"Turkey\", \"asset\":\"PII\"], \"columns\":[{\"name\":\"col1\", \"tags\":[\"residency\":\"Turkey\"]}, {\"name\":\"col2\"}]}}"

	// {"resource":{"name":"file1", "tags":[{"geography":{"name":"Turkey"}}], "columns":[{"name":"col1", "tags":[{"badkey":"Turkey"}]}, {"name":"col2"}]}}
	// resourceGoodInvalidTagKey = "{\"resource\":{\"name\":\"file1\", \"tags\":[residency:\"Turkey\", \"asset\":\"PII\"], \"columns\":[{\"name\":\"col1\", \"tags\":{residency:\"Turkey\"}}, {\"name\":\"col2\"}]}}"
)

func TestCatalogTaxonomy(t *testing.T) {
	ValidateTaxonomy(t, catalogTaxValsName, dataFormatGood, "dataFormatGood", true)
	ValidateTaxonomy(t, catalogTaxValsName, dataFormatBad, "dataFormatBad", false)

	ValidateTaxonomy(t, catalogTaxValsName, dataTypeGood, "dataTypeGood", true)
	ValidateTaxonomy(t, catalogTaxValsName, dataTypeBad, "dataTypeBad", false)

	ValidateTaxonomy(t, catalogTaxStructsName, geographyGood1, "geographyGood1", true)
	ValidateTaxonomy(t, catalogTaxStructsName, geographyGood2, "geographyGood2", true)
	ValidateTaxonomy(t, catalogTaxStructsName, geographyBad1, "geographyBad1", false)

	ValidateTaxonomy(t, catalogTaxStructsName, resourceGoodNameOnly, "resourceGoodNameOnly", true)
	ValidateTaxonomy(t, catalogTaxStructsName, resourceGoodNameGeo, "resourceGoodNameGeo", true)
	ValidateTaxonomy(t, catalogTaxStructsName, resourceGoodCols, "resourceGoodCols", true)
	ValidateTaxonomy(t, catalogTaxStructsName, resourceGoodColsTags, "resourceGoodColsTags", true)

	ValidateTaxonomy(t, catalogTaxStructsName, resourceBadNoName, "resourceBadNoName", false)
	// ValidateTaxonomy(t, catalogTaxStructsName, resourceBadInvalidGeo, "resourceBadInvalidGeo", false)
	// ValidateTaxonomy(t, catalogTaxStructsName, resourceBadInvalidTagVal, "resourceBadInvalidTagVal", false)
	// ValidateTaxonomy(t, catalogTaxStructsName, resourceGoodInvalidTagKey, "resourceBadInvalidTagKey", true)
}
