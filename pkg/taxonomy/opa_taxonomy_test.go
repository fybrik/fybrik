// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"
	tax "github.com/ibm/the-mesh-for-data/config/taxonomy"
)

var (
	OPATaxValsName    = "../../config/taxonomy/opa.structs.schema.json"

governanceRequestGood = "{\"opaGovernanceRequest\": {\"request_context\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\"},\"action\":{\"action_type\":\"read\", \"processingLocation\":\"Turkey\"}, \"resource\":{\"name\":\"file1\"}}}"
governanceRequestBadNoResource = "{\"opaGovernanceRequest\": {\"request_context\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\"},\"action\":{\"action_type\":\"read\", \"processingLocation\":\"Turkey\"}}}"
governanceResponseGood = "{\"opaGovernanceResponse\": {\"resource\":{\"name\":\"file1\"}, \"governance_decision\":\"allow\"}}"
governanceResponseBadNoDecision = "{\"opaGovernanceResponse\": {\"resource\":{\"name\":\"file1\"}}}"

)

func TestOPAInputTaxonomy(t *testing.T) {
	tax.ValidateTaxonomy(t, OPATaxValsName, governanceRequestGood, "governanceRequestGood", true)
	tax.ValidateTaxonomy(t, OPATaxValsName, governanceRequestBadNoResource, "governanceRequestBadNoResource", false)
	tax.ValidateTaxonomy(t, OPATaxValsName, governanceResponseGood, "governanceResponseGood", true)
	tax.ValidateTaxonomy(t, OPATaxValsName, governanceResponseBadNoDecision, "governanceResponseBadNoDecision", false)
}
