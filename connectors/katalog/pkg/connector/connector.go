// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/pkg/model/datacatalog"
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
	b, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("CreateAssetRequest: JSON format: ", string(b))
	log.Println("CreateAssetRequest: ", request)
	output := render.AsCode(request)
	log.Println("CreateAssetRequest - render as code output: ", output)

	// example CreateAssetRequest in JSON
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

	var assetName string
	var namespace string
	if request.DestinationAssetID != "" {
		splittedID := strings.SplitN(string(request.DestinationAssetID), "/", 2)
		if len(splittedID) != 2 {
			errorMessage := fmt.Sprintf("request has an invalid asset ID %s (must be in namespace/name format)", request.ResourceMetadata.Name)
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		}
		namespace, assetName = splittedID[0], splittedID[1]
	} else {
		splittedID := strings.SplitN(string(request.ResourceMetadata.Name), "/", 2)
		if len(splittedID) != 2 {
			errorMessage := fmt.Sprintf("request has an invalid asset ID %s (must be in namespace/name format)", request.ResourceMetadata.Name)
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		}
		namespace, assetName = splittedID[0], splittedID[1]
		// add random string to source asset
		var v [5]int
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 5; i++ {
			v[i] = rand.Intn(100)
		}
		fmt.Println("random integer array:", v) // [0 28 27 62 63]
		delim := "-"
		randomString := strings.Trim(strings.Replace(fmt.Sprint(v), " ", delim, -1), "[]")
		assetName = assetName + "-" + randomString
		log.Println("generated assetName :", assetName)
	}
	log.Println("using assetName :", assetName)

	asset := &v1alpha1.Asset{}
	objectMeta := &v1.ObjectMeta{
		Namespace: namespace,
		Name:      namespace + "/" + assetName,
	}
	asset.ObjectMeta = *objectMeta

	spec := &v1alpha1.AssetSpec{
		SecretRef: v1alpha1.SecretRef{
			Name: request.Credentials,
		},
	}

	reqResourceMetadata, _ := json.Marshal(request.ResourceMetadata)
	err = json.Unmarshal(reqResourceMetadata, &spec.Metadata)
	if err != nil {
		log.Printf("Error during unmarshal of reqResourceMetadata")
		fmt.Println(err)
		return
	}

	reqResourceDetails, _ := json.Marshal(request.ResourceDetails)
	err = json.Unmarshal(reqResourceDetails, &spec.Details)
	if err != nil {
		log.Printf("Error during unmarshal of reqResourceDetails")
		fmt.Println(err)
		return
	}

	asset.Spec = *spec

	b, err = json.Marshal(asset)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("Created Asset: JSON format: ", string(b))
	log.Println("Created Asset: ", asset)
	output = render.AsCode(asset)
	log.Println("Created AssetID - render as code output: ", output)

	log.Printf("Creating Asset in cluster: %s/%s\n", asset.Namespace, asset.Name)
	err = r.client.Create(context.Background(), asset)
	if err != nil {
		log.Printf("Error during create asset")
		fmt.Println(err)
		return
	}

	response := datacatalog.CreateAssetResponse{
		AssetID: namespace + "/" + assetName,
	}

	c.JSON(http.StatusCreated, &response)
}
