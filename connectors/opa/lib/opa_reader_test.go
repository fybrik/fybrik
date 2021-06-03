// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"fmt"
	"os"
	"testing"
	"time"

	tu "github.com/mesh-for-data/mesh-for-data/connectors/opa/testutil"
	"gotest.tools/assert"
)

// Tests  GetPoliciesDecisions in opa-connector.go
// In the test the purpose is set as "marketing". For this purpose in different scenarios connector is mocked and different outputs are obtained.
// In this test the results are  manually synchronised, result of customOpaResponse function should be translated into
// GetExpectedOpaDecisions. Tested here is the functionality of translating opa_result
// format into enforcement decisions format

func TestMainOpaConnector(t *testing.T) {
	timeOutSecs, catalogConnectorURL, opaServerURL := tu.GetEnvironment()
	policyToBeEvaluated := "user_policies"
	applicationContext := tu.GetApplicationContext("marketing")

	srv := NewOpaReader(opaServerURL)
	catalogReader := NewCatalogReader(catalogConnectorURL, timeOutSecs)
	policiesDecisions, err := srv.GetOPADecisions(applicationContext, catalogReader, policyToBeEvaluated)
	assert.NilError(t, err)
	fmt.Println("policiesDecisions returned")
	fmt.Println(policiesDecisions)
	expectedOpaDecisions := tu.GetExpectedOpaDecisions("marketing", applicationContext)
	fmt.Println("expectedOpaDecisions returned")
	fmt.Println(expectedOpaDecisions)

	tu.EnsureDeepEqualDecisions(t, policiesDecisions, expectedOpaDecisions)
}

// TestMain executes the above defined test function.
func TestMain(m *testing.M) {
	fmt.Println("TestMain function called = opa_connector_test ")

	tu.EnvValues["CATALOG_CONNECTOR_URL"] = "localhost:" + "50084"
	tu.EnvValues["OPA_SERVER_URL"] = "localhost:" + "8282"

	go tu.MockCatalogConnector(50084)
	time.Sleep(5 * time.Second)
	go tu.MockOpaServer(8282)
	time.Sleep(5 * time.Second)
	code := m.Run()
	fmt.Println("TestMain function called after Run = opa_connector_test ")
	os.Exit(code)
}
