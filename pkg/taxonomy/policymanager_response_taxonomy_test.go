// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	tax "github.com/ibm/the-mesh-for-data/config/taxonomy"
	"testing"
)

var (
	OPAResponseTaxValsName = "../../config/taxonomy/policymanager_response.structs.schema.json"

	// {\"governance_decision_response\": {\"decision_id\":\"abcde1234\", \"governance_actions\":[{\"actions\":\"\", \"used_policy\":\"policyID112233\"}]}}
	governanceResponseBadNoDecision = "{\"governance_decision_response\": {\"decision_id\":\"abcde1234\", \"result\":[{\"action\": {\"name\":\"\"}, \"used_policy\":\"policyID112233\"}]}}"

	// {"governance_decision_response": {"resource":{"name":"file1"}, "governance_decision":"allow"}}
	governanceResponseGood = "{\"governance_decision_response\": {\"decision_id\":\"abcde1234\", \"result\":[{\"action\": {\"name\":\"RedactColumn\", \"columns\":[\"nameOrig\"]}, \"used_policy\":\"policyID112233\"}]}}"
)

func TestOPAResponseTaxonomy(t *testing.T) {
	tax.ValidateTaxonomy(t, OPAResponseTaxValsName, governanceResponseGood, "governanceResponseGood", true)
	tax.ValidateTaxonomy(t, OPAResponseTaxValsName, governanceResponseBadNoDecision, "governanceResponseBadNoDecision", false)
}
