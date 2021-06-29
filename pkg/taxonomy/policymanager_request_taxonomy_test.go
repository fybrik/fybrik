// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"

	tax "github.com/mesh-for-data/mesh-for-data/config/taxonomy"
)

var (
	OPATaxValsName = "../../charts/m4d/files/taxonomy/policymanager_request.structs.schema.json"

	governanceRequestGood          = "{\"request_context\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\"},\"action\":{\"action_type\":\"read\", \"processingLocation\":\"Turkey\"}, \"resource\":{\"name\":\"file1\"}}"
	governanceRequestBadNoResource = "{\"request_context\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\"},\"action\":{\"action_type\":\"read\", \"processingLocation\":\"Turkey\"}}"
)

func TestOPAInputTaxonomy(t *testing.T) {
	tax.ValidateTaxonomy(t, OPATaxValsName, governanceRequestGood, "governanceRequestGood", true)
	tax.ValidateTaxonomy(t, OPATaxValsName, governanceRequestBadNoResource, "governanceRequestBadNoResource", false)
}
