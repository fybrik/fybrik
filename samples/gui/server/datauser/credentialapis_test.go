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

var _ = Describe("Data User REST API server - credential APIs", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test

	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Credential REST APIs", func() {
		credserverurl := "http://localhost:8080/v1/creds/usercredentials"

		cred1 := "{\"System\": \"Egaria\",\"M4DApplicationID\": \"notebook1\",\"Credentials\": {\"username\": \"user1\"}}"
		cred2 := "{\"System\": \"Egaria\",\"M4DApplicationID\": \"notebook2\",\"Credentials\": {\"username\": \"user2\"}}"
		namespace := "default"
		cred1path := namespace + "/notebook1/Egaria"
		cred2path := namespace + "/notebook2/Egaria"

		It("Store credentials set 1: "+cred1, func() {

			body := strings.NewReader(cred1)
			req, err := http.NewRequest("POST", credserverurl, body)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			defer resp.Body.Close()
		})

		It("Store credentials set 2: "+cred2, func() {

			body := strings.NewReader(cred2)
			req, err := http.NewRequest("POST", credserverurl, body)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			defer resp.Body.Close()
		})

		It("Get credentials set 1: "+cred1path, func() {
			url := credserverurl + "/" + cred1path
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			defer resp.Body.Close()
		})

		It("Get credentials set 2: "+cred2path, func() {
			url := credserverurl + "/" + cred2path
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			defer resp.Body.Close()
		})

		It("Delete credentials set 1: "+cred1path, func() {
			url := credserverurl + "/" + cred1path
			req, err := http.NewRequest("DELETE", url, nil)
			Expect(err).ToNot(HaveOccurred())

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			defer resp.Body.Close()
		})

		It("Delete credentials set 2: "+cred2path, func() {
			url := credserverurl + "/" + cred2path
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

func TestCredAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Credential REST API Suite",
		[]Reporter{printer.NewlineReporter{}})
}
