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

	governanceDecisionNameGood = "{\"governance_decision_name\":\"RedactColumn\"}"
	governanceDecisionNameBad  = "{\"governance_decision_name\":\"xxx\"}"

	protocolGood = "{\"protocol\":\"m4d-arrow-flight\"}"
	protocolBad  = "{\"protocol\":\"xxx\"}"

	// {"transformation": {"name":"mask"}}
	governanceDecisionBaseGood = "{\"governance_decision\": {\"name\":\"RedactColumn\"}}"

	// {"transformation": {"name":"xxx"}}
	governanceDecisionBaseBad = "{\"governance_decision\": {\"name\":\"xxx\"}}"

	// {"transformation": {"name":"mask", "params":["param1", "params2"]}}
	governanceDecisionGoodParams = "{\"governance_decision\": {\"name\":\"RemoveColumn\", \"params\":[\"Column 1\", \"params2\"]}}"

	// {"transformation": {"params":["param1", "params2"]}}
	governanceDecisionBadNoName = "{\"governance_decision\": {\"params\":[\"Column 1\", \"params2\"]}}"
)

func TestModuleTaxonomy(t *testing.T) {
	ValidateTaxonomy(t, ModuleTaxValsName, moduleTypeGood, "moduleTypeGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, moduleTypeBad, "moduleTypeBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, governanceDecisionNameGood, "transformNameGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, governanceDecisionNameBad, "tranformationNameBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, protocolGood, "protocolGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, protocolBad, "protocolBad", false)
	ValidateTaxonomy(t, ModuleTaxStructsName, governanceDecisionBaseGood, "transformationBaseGood", true)
	ValidateTaxonomy(t, ModuleTaxStructsName, governanceDecisionBaseBad, "transformationBaseBad", false)
	ValidateTaxonomy(t, ModuleTaxStructsName, governanceDecisionGoodParams, "transformationGoodParams", true)
	ValidateTaxonomy(t, ModuleTaxStructsName, governanceDecisionBadNoName, "transformationBadNoName", false)
}
