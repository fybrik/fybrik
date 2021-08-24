// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	clients "fybrik.io/fybrik/pkg/connectors/clients"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
)

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Env Variable %v not defined", key)
	}
	log.Printf("Env. variable extracted: %s - %s\n", key, value)
	return value
}

// CatalogReader - Reader struct which has information to read from catalog, this struct does not have any information related to the application context. any request specific info is passed as parameters to functions belonging to this struct.
type CatalogReader struct {
	catalogConnectorAddress string
	timeOut                 int
}

func NewCatalogReader(address string, timeOut int) *CatalogReader {
	return &CatalogReader{catalogConnectorAddress: address, timeOut: timeOut}
}

// return map  datasetID -> metadata of dataset in form of map
func (r *CatalogReader) GetDatasetsMetadataFromCatalog(in *openapiclientmodels.PolicyManagerRequest, creds string) (map[string]interface{}, error) {
	datasetsMetadata := make(map[string]interface{})
	catalogProviderName := getEnv("CATALOG_PROVIDER_NAME")
	datasetID := (in.GetResource()).Name
	if _, present := datasetsMetadata[datasetID]; !present {
		connectionURL := r.catalogConnectorAddress
		connectionTimeout := time.Duration(r.timeOut) * time.Second
		log.Println("creds in GetDatasetsMetadataFromCatalog:", creds)
		log.Println("Create new catalog connection using catalog connector address: ", r.catalogConnectorAddress)

		var dataCatalog clients.DataCatalog
		dataCatalog, err := clients.NewGrpcDataCatalog(catalogProviderName, connectionURL, connectionTimeout)
		if err != nil {
			return nil, fmt.Errorf("connection to external catalog connector failed: %v", err)
		}

		objToSend := &pb.CatalogDatasetRequest{CredentialPath: creds, DatasetId: datasetID}
		info, err := dataCatalog.GetDatasetInfo(context.Background(), objToSend)
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
