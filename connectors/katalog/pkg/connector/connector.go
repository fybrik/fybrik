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
		r.log.Info().Msg(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error during ShouldBindJSON in getAssetInfo "})
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
		if errors.IsNotFound(err) {
			r.log.Info().Msg(err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": "Error: Asset Not Found during getAssetInfo"})
			return
		}
		r.log.Info().Msg(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error during getAssetInfo"})
		return
	}

	secretNamespace := namespace
	if asset.Spec.SecretRef.Namespace != "" {
		secretNamespace = asset.Spec.SecretRef.Namespace
	}

	response := datacatalog.GetAssetResponse{
		ResourceMetadata: asset.Spec.Metadata,
		Details:          asset.Spec.Details,
		Credentials:      vault.PathForReadingKubeSecret(secretNamespace, asset.Spec.SecretRef.Name),
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
// (b) When DestinationAssetID is specified:
//     Then an asset is created with name: <DestinationAssetID>-<Kubernetes Generated Random String>
// (c) When DestinationAssetID is not specified:
//     Then an asset is created with name: fybrik-<Kubernetes Generated Random String>
func (r *Handler) createAsset(c *gin.Context) {
	// Parse request
	var request datacatalog.CreateAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.log.Info().Msg(err.Error())
		r.reportError(c, http.StatusBadRequest, "Error during ShouldBindJSON in createAsset.")
		return
	}

	logging.LogStructure("CreateAssetRequest object received:", request, &r.log, zerolog.DebugLevel, false, false)

	if request.DestinationCatalogID == "" {
		errString := "Invalid DestinationCatalogID in request."
		r.log.Info().Msg(errString)
		r.reportError(c, http.StatusBadRequest, errString)
		return
	}

	secretName, secretNamespace, err := vault.GetKubeSecretDetailsFromVaultPath(request.Credentials)
	if err != nil {
		r.log.Info().Msg(err.Error())
		r.reportError(c, http.StatusInternalServerError, "Error getting kube secret from vaultpath")
		return
	}

	assetPrefix := FybrikAssetPrefix
	if request.DestinationAssetID != "" {
		assetPrefix = request.DestinationAssetID + "-"
	}

	asset := &v1alpha1.Asset{
		ObjectMeta: v1.ObjectMeta{Namespace: request.DestinationCatalogID, GenerateName: assetPrefix},
		Spec: v1alpha1.AssetSpec{
			SecretRef: v1alpha1.SecretRef{Name: secretName, Namespace: secretNamespace},
			Metadata:  request.ResourceMetadata,
			Details:   request.Details,
		},
	}

	logging.LogStructure("Fybrik Asset to be created in Katalog:", asset, &r.log, zerolog.DebugLevel, false, false)

	err = r.client.Create(context.Background(), asset)
	if err != nil {
		r.log.Info().Msg(err.Error())
		r.reportError(c, http.StatusInternalServerError, "Error during create asset.")
		return
	}
	logging.LogStructure("Created Asset: ", asset, &r.log, zerolog.DebugLevel, false, false)

	response := datacatalog.CreateAssetResponse{
		AssetID: asset.ObjectMeta.Name,
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
		r.log.Info().Msg(err.Error())
		r.reportError(c, http.StatusBadRequest, "Error during ShouldBindJSON in deleteAsset ")
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
		r.log.Info().Msg(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while getting asset information"})
		return
	}

	if err := r.client.Delete(context.Background(), asset); err != nil {
		r.log.Info().Msg(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error during deleting asset"})
		return
	}
	response := datacatalog.DeleteAssetResponse{
		Status: "Deletion successful!",
	}
	r.log.Info().Msg(
		"Sending response from Katalog Connector with deleted asset ID: " + string(request.AssetID))

	c.JSON(http.StatusOK, &response)
}

// Enables deletion of assets to katalog.
func (r *Handler) updateAsset(c *gin.Context) {
	// Parse request
	var request datacatalog.UpdateAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.log.Info().Msg(err.Error())
		r.reportError(c, http.StatusBadRequest, "Error during ShouldBindJSON in updateAsset")
		return
	}
	logging.LogStructure("UpdateAssetRequest received:", request, &r.log, zerolog.DebugLevel, false, false)

	splittedID := strings.SplitN(string(request.AssetID), "/", 2)
	if len(splittedID) != 2 {
		errorMessage := fmt.Sprintf("UpdateAssetRequest has an invalid asset ID %s (must be in namespace/name format)", request.AssetID)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}
	namespace, name := splittedID[0], splittedID[1]

	asset := &v1alpha1.Asset{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, asset); err != nil {
		if errors.IsNotFound(err) {
			r.log.Info().Msg(err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": "Error: Asset Not Found during updateAsset"})
			return
		}
		errString := "Error reading asset information"
		r.log.Info().Msg(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errString})
		return
	}

	secretName, secretNamespace, err := vault.GetKubeSecretDetailsFromVaultPath(request.Credentials)
	if err != nil {
		r.log.Info().Msg(err.Error())
		r.reportError(c, http.StatusInternalServerError, "Error obtaining a secret with credentials.")
		return
	}

	asset.Spec.Metadata = request.ResourceMetadata
	asset.Spec.Details = request.Details
	asset.Spec.SecretRef = v1alpha1.SecretRef{Name: secretName, Namespace: secretNamespace}

	if err := r.client.Update(context.Background(), asset); err != nil {
		r.log.Info().Msg(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{" Error ": "Error while updating asset"})
		return
	}
	response := datacatalog.UpdateAssetResponse{
		Status: "Updation successful!",
	}
	r.log.Info().Msg(
		"Sending response from Katalog Connector with updated asset ID: " + string(request.AssetID))

	c.JSON(http.StatusOK, &response)
}
