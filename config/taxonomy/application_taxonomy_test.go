// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"
)

var (
	m4dAppValsName = "application.values.schema.json"

	intentValid   = "{\"intent\":\"Marketing\"}"
	intentInvalid = "{\"intent\":\"whatever\"}"

	roleValid   = "{\"role\":\"Data Scientist\"}"
	roleInvalid = "{\"role\":\"whatever\"}"

	// {"app_info":{"intent":"Marketing", "role":"Data Scientist"}}
	appInfoGood = "{\"app_info\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\"}}"

	// {"app_info":{"intent":"Marketing", "role":"Data Scientist", "x":"Y"}}
	appInfoGoodExtraProps = "{\"app_info\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\", \"x\":\"Y\"}}"

	// {"app_info":{"role":"Data Scientist", "x":"Y"}}
	appInfoBadNoIntent = "{\"app_info\":{\"role\":\"Data Scientist\", \"x\":\"Y\"}}"

	// {"interface":{"protocol":"m4d-arrow-flight", "data_format":"arrow"}}
	interfaceGoodFlight = "{\"interface\":{\"protocol\":\"m4d-arrow-flight\", \"data_format\":\"arrow\"}}"

	// {"interface":{"protocol":"m4d-arrow-flight", "data_format":"parquet"}}
	interfaceBadFlight = "{\"interface\":{\"protocol\":\"m4d-arrow-flight\", \"data_format\":\"parquet\"}}"

	// {"interface":{"protocol":"kafka", "data_format":"json"}}
	interfaceGoodKafka = "{\"interface\":{\"protocol\":\"kafka\", \"data_format\":\"json\"}}"

	// {"interface":{"protocol":"whatever", "data_format":"avro"}}
	interfaceBadKafka = "{\"interface\":{\"protocol\":\"whatever\", \"data_format\":\"avro\"}}"
)

func TestApplicationTaxonomy(t *testing.T) {
	ValidateTaxonomy(t, m4dAppValsName, intentValid, "intentValid", true)
	ValidateTaxonomy(t, m4dAppValsName, intentInvalid, "intentInvalid", false)
	ValidateTaxonomy(t, m4dAppValsName, roleValid, "roleValid", true)
	ValidateTaxonomy(t, m4dAppValsName, roleInvalid, "roleInvalid", false)
	ValidateTaxonomy(t, m4dAppValsName, appInfoGood, "appInfoGood", true)
	ValidateTaxonomy(t, m4dAppValsName, appInfoGoodExtraProps, "appInfoGoodExtraProps", true)
	ValidateTaxonomy(t, m4dAppValsName, appInfoBadNoIntent, "appInfoBadNoIntent", false)
	ValidateTaxonomy(t, m4dAppValsName, interfaceGoodFlight, "interfaceGoodFlight", true)
	ValidateTaxonomy(t, m4dAppValsName, interfaceBadFlight, "interfaceBadFlight", false)
	ValidateTaxonomy(t, m4dAppValsName, interfaceGoodKafka, "interfaceGoodKafka", true)
	ValidateTaxonomy(t, m4dAppValsName, interfaceBadKafka, "interfaceBadKafka", false)
}
