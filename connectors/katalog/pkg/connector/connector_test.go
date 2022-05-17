// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gdexlab/go-render/render"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
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
				Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
					"finance": true,
				}}},
				Columns: []datacatalog.ResourceColumn{
					{
						Name: "c1",
						Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
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
	handler := NewHandler(client)

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
	handler.getAssetInfo(c)
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

func TestCreateAsset(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)
	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	t.Log("Executing TestCreateAsset")

	s3Connection := taxonomy.Connection{
		Name: "s3",
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				"s3": map[string]interface{}{
					"endpoint":   "s3.eu-gb.cloud-object-storage.appdomain.cloud",
					"bucket":     "fybrik-test-bucket",
					"object_key": "small.csv",
				},
			},
		},
	}
	var csvFormat taxonomy.DataFormat = "csv"
	sourceAssetName := "paysim-csv"
	destAssetName := "new-paysim-csv"
	destCatalogID := "fybrik-system"

	// Create a fake request to Katalog connector
	createAssetReq := &datacatalog.CreateAssetRequest{
		DestinationCatalogID: destCatalogID,
		DestinationAssetID:   destAssetName,
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name: sourceAssetName,
			Columns: []datacatalog.ResourceColumn{
				{
					Name: "nameDest",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"PII": true,
					}}},
				},
				{
					Name: "nameOrig",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"SPI": true,
					}}},
				},
			},
		},
		Details: datacatalog.ResourceDetails{
			Connection: s3Connection,
			DataFormat: csvFormat,
		},
		Credentials: "/v1/kubernetes-secrets/dummy-creds?namespace=dummy-namespace2",
	}

	// Create a fake client to mock API calls.
	schema := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(schema)
	client := fake.NewClientBuilder().WithScheme(schema).Build()
	handler := NewHandler(client)

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	requestBytes, err := json.Marshal(createAssetReq)
	g.Expect(err).To(BeNil())
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewBuffer(requestBytes))

	// Call createAsset with the fake request
	handler.createAsset(c)

	t.Run("createAsset", func(t *testing.T) {
		assert.Equal(t, 201, w.Code)

		response := &datacatalog.CreateAssetResponse{}
		err = json.Unmarshal(w.Body.Bytes(), response)
		g.Expect(err).To(BeNil())
		assetName := response.AssetID
		g.Expect(strings.HasPrefix(assetName, destAssetName)).To(BeTrue())

		asset := &v1alpha1.Asset{}
		if err := handler.client.Get(context.Background(),
			types.NamespacedName{Namespace: destCatalogID, Name: assetName}, asset); err != nil {
			t.Log(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		g.Expect(&createAssetReq.ResourceMetadata).To(BeEquivalentTo(&asset.Spec.Metadata))

		// just for logging - start
		b, err := json.Marshal(asset)
		if err != nil {
			fmt.Println(err)
			return
		}
		t.Log("Created Asset in TestCreateAsset : JSON format: ", string(b))
		t.Log("Created Asset in TestCreateAsset : ", asset)
		output := render.AsCode(asset)
		t.Log("Created AssetID in TestCreateAsset - render as code output: ", output)
		t.Log("Completed TestCreateAsset")
		// just for logging - end
	})
}

func TestCreateAssetWthNoDestinationAssetID(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	s3Connection := taxonomy.Connection{
		Name: "s3",
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				"s3": map[string]interface{}{
					"endpoint":   "s3.eu-gb.cloud-object-storage.appdomain.cloud",
					"bucket":     "fybrik-test-bucket",
					"object_key": "small.csv",
				},
			},
		},
	}
	var csvFormat taxonomy.DataFormat = "csv"
	sourceAssetName := "paysim-csv"

	// Create a fake request to Katalog connector
	createAssetReq := &datacatalog.CreateAssetRequest{
		DestinationCatalogID: "fybrik-system",
		DestinationAssetID:   "",
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name: sourceAssetName,
			Columns: []datacatalog.ResourceColumn{
				{
					Name: "nameDest",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"PII": true,
					}}},
				},
				{
					Name: "nameOrig",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"SPI": true,
					}}},
				},
			},
		},
		Details: datacatalog.ResourceDetails{
			Connection: s3Connection,
			DataFormat: csvFormat,
		},
		Credentials: "/v1/kubernetes-secrets/dummy-creds?namespace=dummy-namespace2",
	}

	// Create a fake client to mock API calls.
	schema := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(schema)
	client := fake.NewClientBuilder().WithScheme(schema).Build()
	handler := NewHandler(client)

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	requestBytes, err := json.Marshal(createAssetReq)
	g.Expect(err).To(BeNil())
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewBuffer(requestBytes))

	// Call createAsset with the fake request
	handler.createAsset(c)

	t.Run("createAsset", func(t *testing.T) {
		assert.Equal(t, 201, w.Code)

		response := &datacatalog.CreateAssetResponse{}
		err = json.Unmarshal(w.Body.Bytes(), response)
		g.Expect(err).To(BeNil())
		g.Expect(len(response.AssetID)).To(SatisfyAll(
			BeNumerically(">", len(sourceAssetName))))
		t.Log("response.AssetID: ", response.AssetID)

		assetName := response.AssetID
		namespace := "fybrik-system"
		g.Expect(assetName).Should(HavePrefix(FybrikAssetPrefix))
		t.Log("new asset created with name: ", assetName)

		asset := &v1alpha1.Asset{}
		if err := handler.client.Get(context.Background(),
			types.NamespacedName{Namespace: namespace, Name: assetName}, asset); err != nil {
			t.Log(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// just for checking ResourceMetadata part
		g.Expect(&createAssetReq.ResourceMetadata).To(BeEquivalentTo(&asset.Spec.Metadata))

		// just for logging - start
		b, err := json.Marshal(asset)
		if err != nil {
			fmt.Println(err)
			return
		}
		t.Log("Created Asset in TestCreateAssetWthNoDestinationAssetID : JSON format: ", string(b))
		t.Log("Created Asset in TestCreateAssetWthNoDestinationAssetID : ", asset)
		output := render.AsCode(asset)
		t.Log("Created AssetID in TestCreateAssetWthNoDestinationAssetID - render as code output: ", output)
		t.Log("Completed TestCreateAssetWthNoDestinationAssetID")
		// just for logging - end
	})
}

func TestCreateAndUpdateAsset(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)
	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	t.Log("Executing TestCreateUpdateAsset")

	s3Connection := taxonomy.Connection{
		Name: "s3",
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				"s3": map[string]interface{}{
					"endpoint":   "s3.eu-gb.cloud-object-storage.appdomain.cloud",
					"bucket":     "fybrik-test-bucket",
					"object_key": "small.csv",
				},
			},
		},
	}
	var csvFormat taxonomy.DataFormat = "csv"
	destAssetName := "new-paysim-csv"
	destCatalogID := "fybrik-system"

	// Create a fake request to Katalog connector
	createAssetReq := &datacatalog.CreateAssetRequest{
		DestinationCatalogID: destCatalogID,
		DestinationAssetID:   destAssetName,
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name: destCatalogID + "/" + destAssetName,
			Columns: []datacatalog.ResourceColumn{
				{
					Name: "nameDest",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"PII": true,
					}}},
				},
				{
					Name: "nameOrig",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"SPI": true,
					}}},
				},
			},
		},
		Details: datacatalog.ResourceDetails{
			Connection: s3Connection,
			DataFormat: csvFormat,
		},
		Credentials: "/v1/kubernetes-secrets/dummy-creds?namespace=dummy-namespace2",
	}

	// Create a fake client to mock API calls.
	schema := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(schema)
	client := fake.NewClientBuilder().WithScheme(schema).Build()
	handler := NewHandler(client)

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	requestBytes, err := json.Marshal(createAssetReq)
	g.Expect(err).To(BeNil())
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewBuffer(requestBytes))

	// Call createAsset with the fake request
	handler.createAsset(c)
	assert.Equal(t, 201, w.Code)
	response := &datacatalog.CreateAssetResponse{}
	err = json.Unmarshal(w.Body.Bytes(), response)
	g.Expect(err).To(BeNil())
	g.Expect(response.AssetID).ShouldNot(BeEmpty())
	g.Expect(len(response.AssetID)).To(SatisfyAll(
		BeNumerically(">", len(destAssetName))))
	t.Log("response.AssetID: ", response.AssetID)
	createdAssetID := response.AssetID
	t.Log("createdAssetID: ", createdAssetID)

	s3ConnectionModified := taxonomy.Connection{
		Name: "s3",
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				"s3": map[string]interface{}{
					"endpoint":   "s3.eu-gb.cloud-object-storage.appdomain.cloud",
					"bucket":     "fybrik-test-bucket-changed",
					"object_key": "small.csv",
				},
			},
		},
	}
	// Create a fake request to Katalog connector
	updateAssetReq := &datacatalog.UpdateAssetRequest{
		AssetID: taxonomy.AssetID(destCatalogID + "/" + createdAssetID),
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name: destCatalogID + "/" + createdAssetID,
			Columns: []datacatalog.ResourceColumn{
				{
					Name: "nameDest",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"PII": true,
					}}},
				},
				{
					Name: "nameOrig",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"SPI": true,
					}}},
				},
			},
		},
		Details: datacatalog.ResourceDetails{
			Connection: s3ConnectionModified,
			DataFormat: csvFormat,
		},
		Credentials: "/v1/kubernetes-secrets/dummy-creds?namespace=dummy-namespace2",
	}
	t.Log("updateAssetReq:", updateAssetReq)
	w = httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ = gin.CreateTestContext(w)
	requestBytes, err = json.Marshal(updateAssetReq)
	g.Expect(err).To(BeNil())
	c.Request = httptest.NewRequest(http.MethodPut, "http://localhost/", bytes.NewBuffer(requestBytes))

	// Call updateAsset with the fake request
	handler.updateAsset(c)

	t.Run("updateAsset", func(t *testing.T) {
		assert.Equal(t, 200, w.Code)
		response := &datacatalog.UpdateAssetResponse{}
		err = json.Unmarshal(w.Body.Bytes(), response)
		t.Log("w.Body.String()", w.Body.String())
		g.Expect(err).To(BeNil())

		asset := &v1alpha1.Asset{}
		if err := handler.client.Get(context.Background(),
			types.NamespacedName{Namespace: destCatalogID, Name: createdAssetID}, asset); err != nil {
			t.Log(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		g.Expect(&updateAssetReq.ResourceMetadata).To(BeEquivalentTo(&asset.Spec.Metadata))
		g.Expect(&updateAssetReq.Details).To(BeEquivalentTo(&asset.Spec.Details))

		// just for logging - start
		b, err := json.Marshal(asset)
		if err != nil {
			fmt.Println(err)
			return
		}
		t.Log("Updated Asset in TestCreateUpdateAsset : JSON format: ", string(b))
		t.Log("Updated Asset in TestCreateUpdateAsset : ", asset)
		output := render.AsCode(asset)
		t.Log("Updated AssetID in TestCreateUpdateAsset - render as code output: ", output)
		t.Log("Completed TestCreateUpdateAsset")
		// just for logging - end
	})
}
