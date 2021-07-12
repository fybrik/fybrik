// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"

	tax "github.com/mesh-for-data/mesh-for-data/config/taxonomy"
)

var (
	OPAResponseTaxValsName = "../../charts/m4d/files/taxonomy/policymanager_response.structs.schema.json"

	// {\"governance_decision_response\": {\"decision_id\":\"abcde1234\", \"governance_actions\":[{\"actions\":\"\", \"used_policy\":\"policyID112233\"}]}}
	governanceResponseBadNoDecision = "{\"decision_id\":\"abcde1234\", \"result\":[{\"action\": {\"name\":\"\"}, \"policy\":\"policyID112233\"}]}"

	// {"governance_decision_response": {"resource":{"name":"file1"}, "governance_decision":"allow"}}
	governanceResponseGood = "{\"decision_id\":\"abcde1234\", \"result\":[{\"action\": {\"name\":\"RedactColumn\", \"columns\":[\"nameOrig\"]}, \"policy\":\"policyID112233\"}]}"
)

func TestOPAResponseTaxonomy(t *testing.T) {
	tax.ValidateTaxonomy(t, OPAResponseTaxValsName, governanceResponseGood, "governanceResponseGood", true)
	tax.ValidateTaxonomy(t, OPAResponseTaxValsName, governanceResponseBadNoDecision, "governanceResponseBadNoDecision", false)
}
