// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"fybrik.io/fybrik/pkg/model/storagemanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// test that the implementation agents have been registered successfully
func TestSupportedConnections(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Create a fake client to mock API calls.
	schema := runtime.NewScheme()
	client := fake.NewClientBuilder().WithScheme(schema).Build()
	handler := NewHandler(client)

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/", nil)
	handler.getSupportedStorageTypes(c)
	t.Run("getSupportedStorageTypes", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, w.Code)
		response := &storagemanager.GetSupportedStorageTypesResponse{}
		err := json.Unmarshal(w.Body.Bytes(), response)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(response.ConnectionTypes).To(gomega.HaveLen(2))
		g.Expect(response.ConnectionTypes).To(gomega.ContainElement(taxonomy.ConnectionType("s3")))
		g.Expect(response.ConnectionTypes).To(gomega.ContainElement(taxonomy.ConnectionType("mysql")))
	})
}
