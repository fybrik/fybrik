// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"emperror.dev/errors"

	errors_apimachinery "k8s.io/apimachinery/pkg/api/errors"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/connectors/katalog/utils"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/vault"
	"github.com/gdexlab/go-render/render"
	"github.com/rs/zerolog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// handler.log.Output(zerolog.ConsoleWriter{NoColor: true})
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

func (r *Handler) checkIfAssetDoesNotExistInKatalog(namespace string, name string) error {
	asset := &v1alpha1.Asset{}
	var err error
	if err = r.client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, asset); err != nil {
		if errors_apimachinery.IsNotFound(err) {
			// log.Println("error cause is:", errors.Cause(err))
			errorMessage := "error cause is: " + errors.Cause(err).Error()
			r.log.Error().Msg(errorMessage)
			return nil
		}
	}
	return errors.Wrap(err, "Some other error occurred in checkIfAssetExists")
}

func (r *Handler) createAssetInfo(c *gin.Context) {
	var randomStringLength = 8
	var retriesLimit = 3

	// Parse request
	var request datacatalog.CreateAssetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errorMessage := "Error during ShouldBindJSON in createAssetInfo" + err.Error()
		r.log.Error().Msg(errorMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}

	// just for logging - start
	b, err := json.Marshal(request)
	if err != nil {
		errorMessage := "Error during marshalling request in createAssetInfo" + err.Error()
		r.log.Error().Msg(errorMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}
	r.log.Info().Msg("CreateAssetRequest: JSON format:" + string(b))
	// log.Println("CreateAssetRequest: JSON format: ", string(b))
	r.log.Info().Msg("CreateAssetRequest: " + fmt.Sprintf("%#v", request))
	// log.Println("CreateAssetRequest: ", request)
	output := render.AsCode(request)
	// log.Println("CreateAssetRequest - render as code output: ", output)
	r.log.Info().Msg("CreateAssetRequest - render as code output: " + output)
	// just for logging - end

	if request.DestinationCatalogID == "" {
		errorMessage := "Invalid DestinationCatalogID in request"
		r.log.Error().Msg(errorMessage)
		c.JSON(http.StatusBadRequest, gin.H{"Error": errorMessage})
		return
	}

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
				randomStr, err := utils.GenerateRandomString(randomStringLength)
				if err != nil {
					errorMessage := "Error during GenerateRandomString. Error:" + err.Error()
					// log.Println(errorMessage)
					r.log.Error().Msg(errorMessage)
					// reset assetName
					c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
					return
				}
				assetName = assetName + "-" + randomStr
				err = r.checkIfAssetDoesNotExistInKatalog(namespace, assetName)
				if err == nil {
					break
				} else {
					errorMessage := "Error during checkIfAssetDoesNotExistInCluster. Retrying generation of assetid once more. Error:" + err.Error()
					r.log.Error().Msg(errorMessage)
					// log.Println(errorMessage)
					// reset assetName
					assetName = request.ResourceMetadata.Name
				}
			}
			if i == retriesLimit {
				errorMessage := fmt.Sprintf("Unsuccessful in generating destination asset id with prefix %s. Max retries %d exceeded", request.ResourceMetadata.Name, retriesLimit)
				r.log.Error().Msg(errorMessage)
				c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
				return
			}
		} else {
			// request.ResourceMetadata.Name is null. Then create a random asset name with fybrik-asset as prefix
			namespace = request.DestinationCatalogID
			var i int
			for i := 0; i < retriesLimit; i++ {
				randomStr, err := utils.GenerateRandomString(randomStringLength)
				if err != nil {
					errorMessage := "Error occurred during GenerateRandomString. Error:" + err.Error()
					// log.Println(errorMessage)
					r.log.Error().Msg(errorMessage)
					// reset assetName
					c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
					return
				}
				assetName = "fybrik-asset-" + "-" + randomStr
				err = r.checkIfAssetDoesNotExistInKatalog(namespace, assetName)
				if err == nil {
					break
				} else {
					errorMessage := "Error during checkIfAssetDoesNotExistInCluster. Retrying generation of assetid once more. Error:" + err.Error()
					// log.Println(errorMessage)
					r.log.Error().Msg(errorMessage)
				}
			}
			if i == retriesLimit {
				errorMessage := fmt.Sprintf("Unsuccessful in generating destination asset id. Max retries %d exceeded", retriesLimit)
				r.log.Error().Msg(errorMessage)
				c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
				return
			}
		}
		// log.Println("assetName used with random string generation:", assetName)
		r.log.Info().Msg("assetName used with random string generation:" + assetName)
	}
	// log.Println("assetName used to store the asset :", assetName)
	r.log.Info().Msg("assetName used to store the asset :" + assetName)

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
		errorMessage := "Error during unmarshal of reqResourceMetadata. Error:" + err.Error()
		// log.Println(errorMessage)
		r.log.Error().Msg(errorMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}
	spec.Metadata.Name = assetName

	reqResourceDetails, _ := json.Marshal(request.Details)
	err = json.Unmarshal(reqResourceDetails, &spec.Details)
	if err != nil {
		errorMessage := "Error during unmarshal of reqResourceDetails. Error:" + err.Error()
		// log.Println(errorMessage)
		r.log.Error().Msg(errorMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}

	asset.Spec = *spec

	// just for logging - start
	b, err = json.Marshal(asset)
	if err != nil {
		errorMessage := "Error during Marshal of asset. Error:" + err.Error()
		r.log.Error().Msg(errorMessage)
		// fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r.log.Info().Msg("Created Asset: JSON format: " + string(b))
	// log.Println("Created Asset: JSON format: ", string(b))
	// log.Println("Created Asset: ", asset)
	r.log.Info().Msg("Created Asset: " + fmt.Sprintf("%#v", asset))
	output = render.AsCode(asset)
	// log.Println("Created AssetID - render as code output: ", output)
	r.log.Info().Msg("Created AssetID - render as code output: " + output)
	// just for logging - end

	// log.Printf("Creating Asset in cluster: %s/%s\n", asset.Namespace, asset.Name)
	r.log.Info().Msg("Creating Asset in cluster: " + fmt.Sprintf("%s/%s", asset.Namespace, asset.Name))
	err = r.client.Create(context.Background(), asset)
	if err != nil {
		errorMessage := "Error during create asset. Error:" + err.Error()
		// log.Println(errorMessage)
		r.log.Error().Msg(errorMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}

	response := datacatalog.CreateAssetResponse{
		AssetID: namespace + "/" + assetName,
	}

	c.JSON(http.StatusCreated, &response)
}
