// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

var _ = Describe("Data User REST API server - environment APIs", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test

	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Environment REST API", func() {
		envserverurl := "http://localhost:8080/v1/env/datauserenv"

		sysMap := make(map[string][]string)
		sysMap["Egeria"] = []string{"username"}
		expectedObj := EnvironmentInfo{
			Namespace:       "default",
			Geography:       "US",
			Systems:         sysMap,
			DataSetIDFormat: "{\"ServerName\":\"---\",\"AssetGuid\":\"---\"}",
		}

		It("Query environment info ", func() {
			var err error

			// Call the REST API to get the environment information
			resp, err := http.Get(envserverurl)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Check that the results are what we expected
			// Convert result from json to go struct
			resultObj := EnvironmentInfo{}
			respData, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			respString := string(respData)

			err = json.Unmarshal([]byte(respString), &resultObj)
			Expect(err).ToNot(HaveOccurred())

			Expect(resultObj).To(Equal(expectedObj))

			defer resp.Body.Close()
		})
	})
})

func TestEnvAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"M4DApplication REST API Suite",
		[]Reporter{printer.NewlineReporter{}})
}
