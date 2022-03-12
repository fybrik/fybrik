// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/connectors/katalog/utils"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/vault"
	"github.com/rs/zerolog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Handler struct {
	client kclient.Client
	log    zerolog.Logger
}

type sortByLength []string

// Len implements Len of sort.Interface
func (s sortByLength) Len() int {
	return len(s)
}

// Swap implements Swap of sort.Interface
func (s sortByLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements Less of sort.Interface
func (s sortByLength) Less(i, j int) bool {
	return len(s[i]) > len(s[j])
}

func getLongest(toFind []string) []string {
	// We sort it by length, descending
	sort.Sort(sortByLength(toFind))
	longest := []string{toFind[0]}

	// In case we have more than one element in toFind...
	if len(toFind) > 1 {
		for _, str := range toFind[1:] {
			if len(str) < len(longest[0]) {
				break
			}
			longest = append(longest, str)
		}
	}
	fmt.Println(longest)
	return longest
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

func (r *Handler) generateUniqueAssetName(namespace string, namePrefix string) (string, error) {
	var result v1alpha1.AssetList
	var randomStringLength = 4
	var uniqueAssetName = ""
	err := r.client.List(context.Background(), &result, kclient.InNamespace(namespace))
	if err == nil {
		listOfCandidates := make([]string, 0)
		for i := 0; i < len(result.Items); i++ {
			if strings.Contains(result.Items[i].Spec.Metadata.Name, namePrefix) {
				listOfCandidates = append(listOfCandidates, result.Items[i].Spec.Metadata.Name)
			}
		}
		r.log.Info().Msg("listOfCandidates : " + strings.Join(listOfCandidates, "|"))
		randomStr, err := utils.GenerateRandomString(randomStringLength)
		if err == nil {
			if len(listOfCandidates) > 0 {
				longestArr := getLongest(listOfCandidates)
				r.log.Info().Msg("longestArr : " + strings.Join(longestArr, "|"))
				randIdx := utils.GenerateRandomNumber(0, int64(len(longestArr)))
				r.log.Info().Msg("randIdx : " + fmt.Sprint(randIdx))
				uniqueAssetName = longestArr[randIdx] + "-" + randomStr
			} else {
				// no asset with the given prefix
				uniqueAssetName = namePrefix + "-" + randomStr
			}
			r.log.Info().Msg("uniqueAssetName generated : " + uniqueAssetName)
		} else {
			r.log.Info().Msg("Error during GenerateRandomString: " + err.Error())
		}
	} else {
		r.log.Info().Msg("Error during list operation: " + err.Error())
	}
	return uniqueAssetName, err
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
			assetName, err = r.generateUniqueAssetName(namespace, assetName)
		} else {
			// request.ResourceMetadata.Name is null. Then create a random asset name with fybrik-asset as prefix
			namespace = request.DestinationCatalogID
			assetName, err = r.generateUniqueAssetName(namespace, "fybrik-asset")
		}
		if err != nil {
			errorMessage := "Error during generateUniqueAssetName. Error:" + err.Error()
			r.log.Error().Msg(errorMessage)
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
			return
		}
		r.log.Info().Msg("AssetName used with random string generation:" + assetName)
	}
	r.log.Info().Msg("AssetName used to store the asset :" + assetName)

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
		r.log.Error().Msg(errorMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}
	spec.Metadata.Name = assetName

	reqResourceDetails, _ := json.Marshal(request.Details)
	err = json.Unmarshal(reqResourceDetails, &spec.Details)
	if err != nil {
		errorMessage := "Error during unmarshal of reqResourceDetails. Error:" + err.Error()
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r.log.Info().Msg("Created Asset: JSON format: " + string(b))
	// just for logging - end

	err = r.client.Create(context.Background(), asset)
	if err != nil {
		errorMessage := "Error during create asset. Error:" + err.Error()
		r.log.Error().Msg(errorMessage)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}
	r.log.Info().Msg("Created Asset in cluster: " + fmt.Sprintf("%s/%s", asset.Namespace, asset.Name))

	response := datacatalog.CreateAssetResponse{
		AssetID: namespace + "/" + assetName,
	}

	c.JSON(http.StatusCreated, &response)
}
