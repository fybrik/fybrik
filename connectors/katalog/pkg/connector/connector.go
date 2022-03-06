// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"context"
	"encoding/json"
	"errors"
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
		Details:          asset.Spec.Details,
		Credentials:      vault.PathForReadingKubeSecret(namespace, asset.Spec.SecretRef.Name),
	}

	c.JSON(http.StatusOK, &response)
}

func (r *Handler) createRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	var seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(randomString)
}

func (r *Handler) checkIfAssetExists(namespace string, name string) error {
	asset := &v1alpha1.Asset{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, asset); err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return errors.New("server URL for OPA server must have http or https schema")
	}
	return nil
}

func (r *Handler) createAssetInfo(c *gin.Context) {
	var randomStringLength = 8
	var retriesLimit = 3

	// Parse request
	var request datacatalog.CreateAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// just for logging - start
	b, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("CreateAssetRequest: JSON format: ", string(b))
	log.Println("CreateAssetRequest: ", request)
	output := render.AsCode(request)
	log.Println("CreateAssetRequest - render as code output: ", output)
	// just for logging - end

	var assetName string
	var namespace string
	if request.DestinationAssetID != "" {
		namespace, assetName = request.DestinationCatalogID, request.DestinationAssetID
	} else {
		if request.ResourceMetadata.Name != "" {
			namespace, assetName = request.DestinationCatalogID, request.ResourceMetadata.Name

			var i int
			for i := 0; i < retriesLimit; i++ {
				// add random string to source asset
				randomStr := r.createRandomString(randomStringLength)
				assetName = assetName + "-" + randomStr
				err := r.checkIfAssetExists(namespace, assetName)
				if err == nil {
					break
				}
			}
			if i == retriesLimit {
				errorMessage := fmt.Sprintf("Unsuccessful in generating destination asset id. Max retries %d exceeded", retriesLimit)
				c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
			}
		} else {
			// request.ResourceMetadata.Name is null. Then create a random asset name with fybrik-asset as prefix
			namespace = request.DestinationCatalogID
			var i int
			for i := 0; i < retriesLimit; i++ {
				randomStr := r.createRandomString(randomStringLength)
				assetName = "fybrik-asset-" + "-" + randomStr
				err := r.checkIfAssetExists(namespace, assetName)
				if err == nil {
					break
				}
			}
			if i == retriesLimit {
				errorMessage := fmt.Sprintf("Unsuccessful in generating destination asset id. Max retries %d exceeded", retriesLimit)
				c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
			}
		}
		log.Println("assetName used with random string generation:", assetName)
	}
	log.Println("assetName used to store the asset :", assetName)

	asset := &v1alpha1.Asset{}
	objectMeta := &v1.ObjectMeta{
		Namespace: namespace,
		Name:      assetName,
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
	spec.Metadata.Name = namespace + "/" + assetName

	reqResourceDetails, _ := json.Marshal(request.Details)
	err = json.Unmarshal(reqResourceDetails, &spec.Details)
	if err != nil {
		log.Printf("Error during unmarshal of reqResourceDetails")
		fmt.Println(err)
		return
	}

	asset.Spec = *spec

	// just for logging - start
	b, err = json.Marshal(asset)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("Created Asset: JSON format: ", string(b))
	log.Println("Created Asset: ", asset)
	output = render.AsCode(asset)
	log.Println("Created AssetID - render as code output: ", output)
	// just for logging - end

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
