// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/vault"
)

const (
	FybrikAssetPrefix = "fybrik-"
)

type Handler struct {
	client kclient.Client
	log    zerolog.Logger
}

func NewHandler(client kclient.Client) *Handler {
	handler := &Handler{
		client: client,
		log:    logging.LogInit(logging.CONNECTOR, "katalog-connector"),
	}
	return handler
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
		return
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

func (r *Handler) reportError(c *gin.Context, httpCode int, errorMessage string) {
	r.log.Error().Msg(errorMessage)
	c.JSON(httpCode, gin.H{"error": errorMessage})
}

// Enables writing of assets to katalog. The different flows supported are:
// (a) When DestinationAssetID is specified:
//     Then an asset id is created with name : <DestinationAssetID>
// (b) When DestinationAssetID is not specified:
//     Then an asset is created with name: fybrik-<Kubernetes Generated Random String>
func (r *Handler) createAsset(c *gin.Context) {
	// Parse request
	var request datacatalog.CreateAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.reportError(c, http.StatusBadRequest, "Error during ShouldBindJSON in createAsset"+err.Error())
		return
	}

	logging.LogStructure("CreateAssetRequest object received:", request, &r.log, zerolog.DebugLevel, false, false)

	if request.DestinationCatalogID == "" {
		r.reportError(c, http.StatusBadRequest, "Invalid DestinationCatalogID in request")
		return
	}

	asset := &v1alpha1.Asset{
		ObjectMeta: v1.ObjectMeta{Namespace: request.DestinationCatalogID, Name: request.DestinationAssetID, GenerateName: FybrikAssetPrefix},
		Spec: v1alpha1.AssetSpec{
			SecretRef: v1alpha1.SecretRef{Name: request.Credentials},
			Metadata:  request.ResourceMetadata,
			Details:   request.Details,
		},
	}

	logging.LogStructure("Fybrik Asset to be created in Katalog:", asset, &r.log, zerolog.DebugLevel, false, false)

	err := r.client.Create(context.Background(), asset)
	if err != nil {
		errString := "Error during create asset!"
		if errors.IsAlreadyExists(err) {
			errString = "Asset Already exists!"
		}
		r.reportError(c, http.StatusInternalServerError, errString+" Error: "+err.Error())
		return
	}
	logging.LogStructure("Created Asset: ", asset, &r.log, zerolog.DebugLevel, false, false)

	response := datacatalog.CreateAssetResponse{
		AssetID: request.DestinationCatalogID + "/" + asset.ObjectMeta.Name,
	}
	r.log.Info().Msg(
		"Sending response from Katalog Connector with created asset ID: " + response.AssetID)

	c.JSON(http.StatusCreated, &response)
}

// Enables deletion of assets to katalog.
func (r *Handler) deleteAsset(c *gin.Context) {
	// Parse request
	var request datacatalog.DeleteAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.reportError(c, http.StatusBadRequest, "Error during ShouldBindJSON in deleteAsset"+err.Error())
		return
	}
	logging.LogStructure("DeleteAssetRequest object received:", request, &r.log, zerolog.DebugLevel, false, false)

	splittedID := strings.SplitN(string(request.AssetID), "/", 2)
	if len(splittedID) != 2 {
		errorMessage := fmt.Sprintf("DeleteAssetRequest has an invalid asset ID %s (must be in namespace/name format)", request.AssetID)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}
	namespace, name := splittedID[0], splittedID[1]

	asset := &v1alpha1.Asset{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := r.client.Delete(context.Background(), asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := datacatalog.DeleteAssetResponse{
		Status: "Deletion successful!",
	}
	r.log.Info().Msg(
		"Sending response from Katalog Connector with deleted asset ID: " + string(request.AssetID))

	c.JSON(http.StatusOK, &response)
}
