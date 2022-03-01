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

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
	"github.com/gdexlab/go-render/render"
	"k8s.io/apimachinery/pkg/types"
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
		g.Expect(&response.ResourceDetails).To(BeEquivalentTo(&asset.Spec.Details))
		g.Expect(&response.ResourceMetadata).To(BeEquivalentTo(&asset.Spec.Metadata))
		g.Expect(response.Credentials).To(BeEquivalentTo("/v1/kubernetes-secrets/creds-demo-asset?namespace=demo"))
	})
}

func TestCreateAssetInfo(t *testing.T) {
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
	var sourceAssetName string = "fybrik-system/paysim-csv"
	var destAssetName string = "fybrik-system/new-paysim-csv"

	// Create a fake request to Katalog connector
	createAssetReq := &datacatalog.CreateAssetRequest{
		DestinationCatalogID: "testcatalogid",
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
		ResourceDetails: datacatalog.ResourceDetails{
			Connection: s3Connection,
			DataFormat: csvFormat,
		},
		Credentials: "http://fybrik-system:8200/v1/kubernetes-secrets/wkc-creds?namespace=cp4d",
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

	// Call createAssetInfo with the fake request
	handler.createAssetInfo(c)

	t.Run("createAssetInfo", func(t *testing.T) {
		assert.Equal(t, 201, w.Code)

		response := &datacatalog.CreateAssetResponse{}
		err = json.Unmarshal(w.Body.Bytes(), response)
		g.Expect(err).To(BeNil())
		g.Expect(&response.AssetID).To(BeEquivalentTo(&destAssetName))

		splittedID := strings.SplitN(string(destAssetName), "/", 2)
		if len(splittedID) != 2 {
			errorMessage := fmt.Sprintf("request has an invalid destAssetName %s (must be in namespace/name format)", destAssetName)
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		}
		namespace, assetName := splittedID[0], splittedID[1]

		asset := &v1alpha1.Asset{}
		if err := handler.client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: namespace + "/" + assetName}, asset); err != nil {
			t.Log(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		g.Expect(&createAssetReq.ResourceMetadata).To(BeEquivalentTo(&asset.Spec.Metadata))

		b, err := json.Marshal(asset)
		if err != nil {
			fmt.Println(err)
			return
		}
		t.Log("Created Asset in TestCreateAssetInfo : JSON format: ", string(b))
		t.Log("Created Asset in TestCreateAssetInfo : ", asset)
		output := render.AsCode(asset)
		t.Log("Created AssetID in TestCreateAssetInfo - render as code output: ", output)
	})
}
