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

	actionNameColumnsGood = "{\"action_columns\":\"RedactColumn\"}"
	actionNameColumnsBad  = "{\"action_columns\":\"DenyWriting\"}"

	actionNameDatasetGood = "{\"action_dataset\":\"DenyAccess\"}"
	actionNameDatasetBad  = "{\"action_dataset\":\"RedactColumn\"}"

	protocolGood = "{\"protocol\":\"m4d-arrow-flight\"}"
	protocolBad  = "{\"protocol\":\"xxx\"}"

	actionNameGood = "{\"action\": {\"name\":\"DenyAccess\"}}"
	actionNameBad  = "{\"action\": {\"name\":\"xxx\"}}"

	actionGoodRequiredField    = "{\"action\": {\"name\":\"RedactColumn\", \"columns\":[\"nameOrig\"]}}"
	actionMissingRequiredField = "{\"action\": {\"name\":\"RemoveColumn\"}}"
)

func TestModuleTaxonomy(t *testing.T) {
	ValidateTaxonomy(t, ModuleTaxValsName, moduleTypeGood, "moduleTypeGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, moduleTypeBad, "moduleTypeBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, actionNameColumnsGood, "actionNameColumnsGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, actionNameColumnsBad, "actionNameColumnsBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, actionNameDatasetGood, "actionNameDatasetGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, actionNameDatasetBad, "actionNameDatasetBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, protocolGood, "protocolGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, protocolBad, "protocolBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, actionNameGood, "actionNameGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, actionNameBad, "actionNameBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, actionGoodRequiredField, "actionGoodRequiredField", true)
	ValidateTaxonomy(t, ModuleTaxValsName, actionMissingRequiredField, "actionMissingRequiredField", false)
}
