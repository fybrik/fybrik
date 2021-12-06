// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	clients "fybrik.io/fybrik/pkg/connectors/clients"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	datacatalogTaxonomyModels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
)

// CatalogReader - Reader struct which has information to read from catalog, this struct does not have any information related to the application context. any request specific info is passed as parameters to functions belonging to this struct.
type CatalogReader struct {
	DataCatalog *clients.DataCatalog
}

func NewCatalogReader(dataCatalog *clients.DataCatalog) *CatalogReader {
	return &CatalogReader{DataCatalog: dataCatalog}
}

// return map  datasetID -> metadata of dataset in form of map
func (r *CatalogReader) GetDatasetsMetadataFromCatalog(in *taxonomymodels.PolicyManagerRequest, creds string) (map[string]interface{}, error) {
	datasetsMetadata := make(map[string]interface{})
	datasetID := (in.GetResource()).Name
	if _, present := datasetsMetadata[datasetID]; !present {
		// objToSend := &pb.CatalogDatasetRequest{CredentialPath: creds, DatasetId: datasetID}
		objToSend := datacatalogTaxonomyModels.DataCatalogRequest{AssetID: datasetID}

		info, err := (*r.DataCatalog).GetAssetInfo(&objToSend, creds)
		// info, err := (*r.DataCatalog).GetDatasetInfo(context.Background(), objToSend)
		if err != nil {
			return nil, err
		}

		log.Printf("Received Response from External Catalog Connector for  dataSetID: %s\n", datasetID)
		log.Printf("Response received from External Catalog Connector is given below:")
		responseBytes, errJSON := json.MarshalIndent(info, "", "\t")
		if errJSON != nil {
			return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
		}
		log.Print(string(responseBytes))
		metadataMap := make(map[string]interface{})
		err = json.Unmarshal(responseBytes, &metadataMap)
		if err != nil {
			return nil, fmt.Errorf("error in unmarshalling responseBytes (datasetID = %s): %v", datasetID, err)
		}
		datasetsMetadata[datasetID] = metadataMap
	}

	return datasetsMetadata, nil
}

func (r *CatalogReader) GetDatasetMetadata(ctx *context.Context, client pb.DataCatalogServiceClient, datasetID string, creds string) (map[string]interface{}, error) {
	objToSend := &pb.CatalogDatasetRequest{CredentialPath: creds, DatasetId: datasetID}
	log.Printf("Sending request to External Catalog Connector: datasetID = %s", datasetID)
	info, err := client.GetDatasetInfo(*ctx, objToSend)
	if err != nil {
		return nil, fmt.Errorf("error sending data to External Catalog Connector (datasetID = %s): %v", datasetID, err)
	}

	log.Printf("Received Response from External Catalog Connector for  dataSetID: %s\n", datasetID)
	log.Printf("Response received from External Catalog Connector is given below:")
	responseBytes, errJSON := json.MarshalIndent(info, "", "\t")
	if errJSON != nil {
		return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
	}
	log.Print(string(responseBytes))
	metadataMap := make(map[string]interface{})
	err = json.Unmarshal(responseBytes, &metadataMap)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshalling responseBytes (datasetID = %s): %v", datasetID, err)
	}

	return metadataMap, nil
}
