// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/vault"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
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

func (r *Handler) reportError(errorMessage string, c *gin.Context, httpCode int) {
	r.log.Error().Msg(errorMessage)
	c.JSON(httpCode, gin.H{"error": errorMessage})
}

// Enables writing of assets to katalog. The different flows supported are:
// (a) When DestinationAssetID is specified:
//     Then a destination asset id is created with name : <DestinationAssetID>
// (b) When DestinationAssetID is not specified but ResourceMetadata.Name of source asset is specified:
//     Then an asset is created with name: ResourceMetadata.Name-<RANDOMSTRING_LENGTH_4>
// (c) When DestinationAssetID and ResourceMetadata.Name of source asset are not specified:
//     Then an asset is created with name: fybrik-asset-<RANDOMSTRING_LENGTH_4>
func (r *Handler) createAssetInfo(c *gin.Context) {
	// Parse request
	var request datacatalog.CreateAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.reportError("Error during ShouldBindJSON in createAssetInfo"+err.Error(), c, http.StatusBadRequest)
		return
	}

	logging.LogStructure("CreateAssetRequest object received:", request, &r.log, false, false)

	if request.DestinationCatalogID == "" {
		r.reportError("Invalid DestinationCatalogID in request", c, http.StatusBadRequest)
		return
	}

	// var assetName string
	// var err error
	// namespace := request.DestinationCatalogID
	// if request.DestinationAssetID != "" {
	// 	assetName = request.DestinationAssetID
	// } else {
	// 	if request.ResourceMetadata.Name != "" {
	// 		assetName, err = utils.GenerateUniqueAssetName(
	// 			namespace, request.ResourceMetadata.Name, &r.log, r.client)
	// 	} else {
	// 		assetName, err = utils.GenerateUniqueAssetName(namespace, "fybrik-asset", &r.log, r.client)
	// 	}
	// 	if err != nil {
	// 		r.reportError("Error during generateUniqueAssetName. Error: "+err.Error(), c, http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	var asset *v1alpha1.Asset
	var err error
	assetNamePrefix := ""
	namespace := request.DestinationCatalogID
	if request.DestinationAssetID != "" {
		asset = &v1alpha1.Asset{
			ObjectMeta: v1.ObjectMeta{Namespace: namespace, Name: request.DestinationAssetID},
			Spec: v1alpha1.AssetSpec{
				SecretRef: v1alpha1.SecretRef{Name: request.Credentials},
				Metadata:  request.ResourceMetadata,
				Details:   request.Details,
			},
		}
	} else {
		if request.ResourceMetadata.Name != "" {
			assetNamePrefix = request.ResourceMetadata.Name
		} else {
			assetNamePrefix = "fybrik-asset"
		}
		r.log.Info().Msg("Using assetNamePrefix : " + assetNamePrefix)
		asset = &v1alpha1.Asset{
			ObjectMeta: v1.ObjectMeta{Namespace: namespace, GenerateName: assetNamePrefix},
			Spec: v1alpha1.AssetSpec{
				SecretRef: v1alpha1.SecretRef{Name: request.Credentials},
				Metadata:  request.ResourceMetadata,
				Details:   request.Details,
			},
		}
	}
	logging.LogStructure("Fybrik Asset to be created in Katalog:", asset, &r.log, false, false)

	err = r.client.Create(context.Background(), asset)
	if err != nil {
		r.reportError("Error during create asset. Error:"+err.Error(), c, http.StatusInternalServerError)
		return
	}
	logging.LogStructure("Created Asset: ", asset, &r.log, false, false)

	if assetNamePrefix != "" {
		asset.Spec.Metadata.Name = asset.ObjectMeta.Name
		err = r.client.Update(context.Background(), asset)
		if err != nil {
			r.reportError("Error during update asset. Error:"+err.Error(), c, http.StatusInternalServerError)
			return
		}

		err := wait.Poll(500*time.Millisecond, 5*time.Second, func() (bool, error) {
			err := r.client.Update(context.Background(), asset)
			if err != nil && errors.IsConflict(err) {
				r.log.Info().Msg("Error updating asset :(will retry) ")
				return false, nil
			}
			return true, err
		})
		if err != nil {
			r.reportError("Update asset failed after retrying for 5 seconds. Error:"+err.Error(), c, http.StatusInternalServerError)
			return
		}
		logging.LogStructure("Updated Asset: ", asset, &r.log, false, false)
	}

	response := datacatalog.CreateAssetResponse{
		AssetID: namespace + "/" + asset.ObjectMeta.Name,
	}
	r.log.Info().Msg(
		"Sending response from Katalog Connector with created asset ID: " + response.AssetID)

	c.JSON(http.StatusCreated, &response)
}
