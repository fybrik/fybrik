// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestGetAssetInfo(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	// Create fake Asset
	asset := &v1alpha1.Asset{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "demo",
			Name:      "demo-asset",
		},
		Spec: v1alpha1.AssetSpec{
			SecretRef: v1alpha1.SecretRef{
				Name: "creds-demo-asset",
			},
			Details: datacatalog.ResourceDetails{
				Connection: taxonomy.Connection{
					Name: "dummy",
				},
			},
			Metadata: datacatalog.ResourceMetadata{
				Name:      "demoAsset",
				Owner:     "Alice",
				Geography: "us-south",
				Tags: taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
					"finance": true,
				}}},
				Columns: []datacatalog.ResourceColumn{
					{
						Name: "c1",
						Tags: taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
							"PII": true,
						}}},
					},
				},
			},
		},
	}

	// Create a fake client to mock API calls.
	schema := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(schema)
	client := fake.NewClientBuilder().WithScheme(schema).WithObjects(asset).Build()
	controller := NewConnectorController(client)

	// Create a fake request to Katalog connector
	request := &datacatalog.GetAssetRequest{
		AssetID:       "demo/demo-asset",
		OperationType: datacatalog.READ,
	}
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	requestBytes, err := json.Marshal(request)
	g.Expect(err).To(BeNil())
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewBuffer(requestBytes))

	// Call getAssetInfo with the fake request
	controller.getAssetInfo(c)
	t.Run("getAssetInfo", func(t *testing.T) {
		assert.Equal(t, 200, w.Code)

		response := &datacatalog.GetAssetResponse{}
		err = json.Unmarshal(w.Body.Bytes(), response)
		g.Expect(err).To(BeNil())
		g.Expect(&response.Details).To(BeEquivalentTo(&asset.Spec.Details))
		g.Expect(&response.ResourceMetadata).To(BeEquivalentTo(&asset.Spec.Metadata))
		g.Expect(response.Credentials).To(BeEquivalentTo("/v1/kubernetes-secrets/creds-demo-asset?namespace=demo"))

	})
}
