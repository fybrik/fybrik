// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"fmt"
	"os"
	"testing"
	"time"

	tu "fybrik.io/fybrik/connectors/open_policy_agent/testutil"
	clients "fybrik.io/fybrik/pkg/connectors/clients"
	connectors "fybrik.io/fybrik/pkg/connectors/clients"
	"github.com/hashicorp/go-retryablehttp"
	"gotest.tools/assert"
)

// Tests  GetPoliciesDecisions in opa-connector.go
// In the test the purpose is set as "marketing". For this purpose in different scenarios connector is mocked and different outputs are obtained.
// In this test the results are  manually synchronised, result of customOpaResponse function should be translated into
// GetExpectedOpaDecisions. Tested here is the functionality of translating opa_result
// format into enforcement decisions format

func TestMainOpaConnector(t *testing.T) {
	timeOutSecs, catalogConnectorURL, opaServerURL, catalogProviderName := tu.GetEnvironment()
	os.Setenv("CATALOG_PROVIDER_NAME", catalogProviderName)
	defer os.Unsetenv("CATALOG_PROVIDER_NAME")

	policyToBeEvaluated := "user_policies"
	applicationContext := tu.GetApplicationContext("marketing")

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	standardClient := retryClient.HTTPClient // *http.Client

	srv := NewOpaReader(opaServerURL, standardClient)

	connectionTimeout := time.Duration(timeOutSecs) * time.Second
	dataCatalog, err := clients.NewGrpcDataCatalog(catalogProviderName, catalogConnectorURL, connectionTimeout)
	assert.NilError(t, err)

	catalogReader := NewCatalogReader(&dataCatalog)
	policyManagerReq, creds, err := connectors.ConvertGrpcReqToOpenAPIReq(applicationContext)
	assert.NilError(t, err)

	policyManagerResp, err := srv.GetOPADecisions(policyManagerReq, creds, catalogReader, policyToBeEvaluated)
	assert.NilError(t, err)

	datasets := applicationContext.GetDatasets()
	op := datasets[0].GetOperation()
	datasetID := datasets[0].GetDataset().GetDatasetId()
	policiesDecisions, err := connectors.ConvertOpenAPIRespToGrpcResp(&policyManagerResp, datasetID, op)
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

	tu.EnvValues["CATALOG_CONNECTOR_URL"] = "localhost:" + "50085"
	tu.EnvValues["OPA_SERVER_URL"] = "localhost:" + "8383"
	tu.EnvValues["CATALOG_PROVIDER_NAME"] = "dummy_catalog"

	go tu.MockCatalogConnector(50085)
	time.Sleep(5 * time.Second)
	go tu.MockOpaServer(8383)
	time.Sleep(5 * time.Second)
	code := m.Run()
	fmt.Println("TestMain function called after Run = opa_connector_test ")
	os.Exit(code)
}
