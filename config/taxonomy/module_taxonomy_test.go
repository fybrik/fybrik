// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"
)

var (
	ModuleTaxStructsName = "module.structs.schema.json"
	ModuleTaxValsName    = "module.values.schema.json"

	moduleTypeGood = "{\"module_type\":\"read\"}"
	moduleTypeBad  = "{\"module_type\":\"xxx\"}"

	transformNameGood    = "{\"transformation_name\":\"mask\"}"
	tranformationNameBad = "{\"transformation_name\":\"xxx\"}"

	dataLevelNameGood = "{\"data_level_name\":\"COLUMN\"}"
	dataLevelNameBad  = "{\"data_level_name\":\"XXX\"}"

	dataLevelCodeGood = "{\"data_level_code\":0}"
	dataLevelCodeBad  = "{\"data_level_code\":5}"

	protocolGood = "{\"protocol\":\"m4d-arrow-flight\"}"
	protocolBad  = "{\"protocol\":\"xxx\"}"

	// {"transformation": {"name":"mask"}}
	transformationBaseGood = "{\"transformation\": {\"name\":\"mask\"}}"

	// {"transformation": {"name":"xxx"}}
	transformationBaseBad = "{\"transformation\": {\"name\":\"xxx\"}}"

	// {"transformation": {"name":"mask", "params":["param1", "params2"]}}
	transformationGoodParams = "{\"transformation\": {\"name\":\"mask\", \"params\":[\"param1\", \"params2\"]}}"

	// {"transformation": {"params":["param1", "params2"]}}
	transformationBadNoName = "{\"transformation\": {\"params\":[\"param1\", \"params2\"]}}"

	// {"transformation": {"name":"mask", "level": {"name":"DATASET", "value":1},"params":["param1", "params2"]}}
	transformationGoodLevel = "{\"transformation\": {\"name\":\"mask\", \"level\": {\"name\":\"DATASET\", \"value\":1},\"params\":[\"param1\", \"params2\"]}}"

	// {"transformation": {"name":"mask", "level": {"name":"DATASET", "value":10},"params":["param1", "params2"]}}
	transformationBadLevel = "{\"transformation\": {\"name\":\"mask\", \"level\": {\"name\":\"DATASET\", \"value\":10},\"params\":[\"param1\", \"params2\"]}}"
)

func TestModuleTaxonomy(t *testing.T) {
	ValidateTaxonomy(t, ModuleTaxValsName, moduleTypeGood, "moduleTypeGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, moduleTypeBad, "moduleTypeBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, transformNameGood, "transformNameGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, tranformationNameBad, "tranformationNameBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, dataLevelNameGood, "dataLevelNameGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, dataLevelNameBad, "dataLevelNameBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, dataLevelCodeGood, "dataLevelCodeGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, dataLevelCodeBad, "dataLevelCodeBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, protocolGood, "protocolGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, protocolBad, "protocolBad", false)

	ValidateTaxonomy(t, ModuleTaxStructsName, transformationBaseGood, "transformationBaseGood", true)
	ValidateTaxonomy(t, ModuleTaxStructsName, transformationBaseBad, "transformationBaseBad", false)
	ValidateTaxonomy(t, ModuleTaxStructsName, transformationGoodParams, "transformationGoodParams", true)
	ValidateTaxonomy(t, ModuleTaxStructsName, transformationBadNoName, "transformationBadNoName", false)
	ValidateTaxonomy(t, ModuleTaxStructsName, transformationGoodLevel, "transformationGoodLevel", true)
	ValidateTaxonomy(t, ModuleTaxStructsName, transformationBadLevel, "transformationBadLevel", false)
}
