// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

var _ = Describe("Data User REST API server", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test

	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("M4DApplication REST APIs", func() {
		dmaserverurl := "http://localhost:8080/v1/dma/m4dapplication"

		dm1 := "{\"apiVersion\": \"app.m4d.ibm.com/v1alpha1\",\"kind\": \"M4DApplication\",\"metadata\": {\"name\": \"unittest-notebook1\"},\"spec\": {\"selector\": {\"matchLabels\":{\"app\": \"notebook1\"}},\"appInfo\": {\"purpose\": \"fraud-detection\",\"role\": \"Security\"}, \"data\": [{\"dataSetID\": \"2d1b5352-1fbf-439b-8bb0-c1967ac484b9\",\"ifDetails\": {\"protocol\": \"s3\",\"dataformat\": \"parquet\"}}]}}"
		dm1name := "unittest-notebook1"
		dm2 := "{\"apiVersion\": \"app.m4d.ibm.com/v1alpha1\",\"kind\": \"M4DApplication\",\"metadata\": {\"name\": \"unittest-notebook2\"},\"spec\": {\"selector\": {\"matchLabels\":{\"app\": \"notebook2\"}},\"appInfo\": {\"purpose\": \"fraud-detection\",\"role\": \"Security\"}, \"data\": [{\"dataSetID\": \"2d1b5352-1fbf-439b-8bb0-c1967ac484b9\",\"ifDetails\": {\"protocol\": \"s3\",\"dataformat\": \"parquet\"}}]}}"
		dm2name := "unittest-notebook2"

		It("Create M4DApplication "+dm1name, func() {
			var err error

			// Call the REST API to create a new M4DApplication CRD
			body := strings.NewReader(dm1)
			req, err := http.NewRequest("POST", dmaserverurl, body)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			defer resp.Body.Close()
		})

		It("Create M4DApplication "+dm2name, func() {
			body := strings.NewReader(dm2)
			req, err := http.NewRequest("POST", dmaserverurl, body)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			defer resp.Body.Close()
		})

		It("Get a list of all M4DApplications", func() {
			resp, err := http.Get(dmaserverurl)
			Expect(err).ToNot(HaveOccurred())

			// TODO - Fix expectations to check results
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			defer resp.Body.Close()
		})

		It("Get a specific M4DApplication - "+dm1name, func() {
			url := dmaserverurl + "/" + dm1name
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			// TODO - Fix expectations to check results
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			defer resp.Body.Close()
		})

		It("Delete M4DApplication "+dm1name, func() {
			// Call the REST API to delete an existing M4DApplication CRD
			url := dmaserverurl + "/" + dm1name
			req, err := http.NewRequest("DELETE", url, nil)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			defer resp.Body.Close()
		})

		It("Delete M4DApplication "+dm2name, func() {
			// Call the REST API to delete an existing M4DApplication CRD
			url := dmaserverurl + "/" + dm2name
			req, err := http.NewRequest("DELETE", url, nil)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			defer resp.Body.Close()
		})

	})
})

func TestDMAAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"M4DApplication REST API Suite",
		[]Reporter{printer.NewlineReporter{}})
}
