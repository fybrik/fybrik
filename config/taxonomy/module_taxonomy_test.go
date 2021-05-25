// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"
)

var (
	// ModuleTaxStructsName = "module.structs.schema.json"
	ModuleTaxValsName = "module.values.schema.json"

	moduleTypeGood = "{\"module_type\":\"read\"}"
	moduleTypeBad  = "{\"module_type\":\"xxx\"}"

	// governanceDecisionNameGood = "{\"action_name\":\"RedactColumn\"}"
	// governanceDecisionNameBad  = "{\"action_name\":\"xxx\"}"

	protocolGood = "{\"protocol\":\"m4d-arrow-flight\"}"
	protocolBad  = "{\"protocol\":\"xxx\"}"

	// {"transformation": {"name":"mask"}}
	// governanceDecisionBaseGood = "{\"action\": {\"name\":\"RedactColumn\"}}"

	// {"transformation": {"name":"xxx"}}
	// governanceDecisionBaseBad = "{\"action\": {\"name\":\"xxx\"}}"

	// {"transformation": {"name":"mask", "params":["param1", "params2"]}}
	// governanceDecisionGoodParams = "{\"action\": {\"name\":\"RemoveColumn\", \"params\":[\"Column 1\", \"params2\"]}}"

	// {"transformation": {"params":["param1", "params2"]}}
	// governanceDecisionBadNoName = "{\"action\": {\"params\":[\"Column 1\", \"params2\"]}}"
)

func TestModuleTaxonomy(t *testing.T) {
	ValidateTaxonomy(t, ModuleTaxValsName, moduleTypeGood, "moduleTypeGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, moduleTypeBad, "moduleTypeBad", false)
	// ValidateTaxonomy(t, ModuleTaxValsName, governanceDecisionNameGood, "governanceDecisionNameGood", true)
	// ValidateTaxonomy(t, ModuleTaxValsName, governanceDecisionNameBad, "governanceDecisionNameBad", false)
	ValidateTaxonomy(t, ModuleTaxValsName, protocolGood, "protocolGood", true)
	ValidateTaxonomy(t, ModuleTaxValsName, protocolBad, "protocolBad", false)
	// ValidateTaxonomy(t, ModuleTaxStructsName, governanceDecisionBaseGood, "governanceDecisionBaseGood", true)
	// ValidateTaxonomy(t, ModuleTaxStructsName, governanceDecisionBaseBad, "governanceDecisionBaseBad", false)
	// ValidateTaxonomy(t, ModuleTaxStructsName, governanceDecisionGoodParams, "governanceDecisionGoodParams", true)
	// ValidateTaxonomy(t, ModuleTaxStructsName, governanceDecisionBadNoName, "governanceDecisionBadNoName", false)
}
