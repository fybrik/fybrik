// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/vault"
	"github.com/gdexlab/go-render/render"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Handler struct {
	client kclient.Client
}

func NewHandler(client kclient.Client) *Handler {
	return &Handler{
		client: client,
	}
}

func (r *Handler) getAssetInfo(c *gin.Context) {
	// Parse request
	var request datacatalog.GetAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	splittedID := strings.SplitN(string(request.AssetID), "/", 2)
	if len(splittedID) != 2 {
		errorMessage := fmt.Sprintf("request has an invalid asset ID %s (must be in namespace/name format)", request.AssetID)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
	}
	namespace, name := splittedID[0], splittedID[1]

	asset := &v1alpha1.Asset{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := datacatalog.GetAssetResponse{
		ResourceMetadata: asset.Spec.Metadata,
		ResourceDetails:  asset.Spec.Details,
		Credentials:      vault.PathForReadingKubeSecret(namespace, asset.Spec.SecretRef.Name),
	}

	c.JSON(http.StatusOK, &response)
}

func (r *Handler) createAssetInfo(c *gin.Context) {
	// Parse request
	var request datacatalog.CreateAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("CreateAssetRequest: ", request)
	b, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("CreateAssetRequest: JSON format: ", string(b))

	// example output in JSON
	// {
	// 	"destinationCatalogID": "testcatalogid",
	// 	"resourceMetadata": {
	// 		"name": "demoAsset",
	// 		"owner": "Alice",
	// 		"geography": "us-south",
	// 		"tags": {
	// 			"finance": true
	// 		},
	// 		"columns": [
	// 			{
	// 				"name": "c1",
	// 				"tags": {
	// 					"PII": true
	// 				}
	// 			}
	// 		]
	// 	},
	// 	"resourceDetails": {
	// 		"connection": {
	// 			"name": "s3"
	// 		}
	// 	},
	// 	"credentials": "http://fybrik-system:8200/v1/kubernetes-secrets/wkc-creds?namespace=cp4d"
	// }

	asset := &v1alpha1.Asset{}
	objectMeta := &v1.ObjectMeta{
		Namespace: "fybrik-system",
		Name:      "demo-asset123",
	}
	asset.ObjectMeta = *objectMeta
	spec := &v1alpha1.AssetSpec{
		SecretRef: v1alpha1.SecretRef{
			Name: request.Credentials,
		},
		Details: datacatalog.ResourceDetails{
			Connection: taxonomy.Connection{
				Name: request.ResourceDetails.Connection.Name,
			},
		},
		Metadata: datacatalog.ResourceMetadata{
			Name:      request.ResourceMetadata.Name,
			Owner:     request.ResourceMetadata.Owner,
			Geography: request.ResourceMetadata.Geography,
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
	}
	asset.Spec = *spec
	log.Printf("Creating Asset for first ever time %s/%s\n", asset.Namespace, asset.Name)
	err = r.client.Create(context.Background(), asset)
	if err != nil {
		log.Printf("Error during create asset")
		fmt.Println(err)
		return
	}
	log.Println("CreateAssetRequest: ", request)
	output := render.AsCode(request)
	log.Println("CreateAssetRequest - render as code output: ", output)

	response := datacatalog.CreateAssetResponse{
		AssetID: "test",
	}

	c.JSON(http.StatusOK, &response)
}
