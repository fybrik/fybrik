// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/vault"
	"github.com/gdexlab/go-render/render"
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
	output := render.AsCode(request)
	log.Println("CreateAssetRequest - render as code output: ", output)

	response := datacatalog.CreateAssetResponse{
		AssetID: "test",
	}

	c.JSON(http.StatusOK, &response)
}
