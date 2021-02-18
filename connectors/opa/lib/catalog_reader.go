// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

// CatalogReader - Reader struct which has information to read from catalog, this struct does not have any information related to the application context. any request specific info is passed as parameters to functions belonging to this struct.
type CatalogReader struct {
	catalogConnectorAddress string
	timeOut                 int
}

func NewCatalogReader(address string, timeOut int) *CatalogReader {
	return &CatalogReader{catalogConnectorAddress: address, timeOut: timeOut}
}

// return map  datasetID -> metadata of dataset in form of map
func (r *CatalogReader) GetDatasetsMetadataFromCatalog(in *pb.ApplicationContext) (map[string]interface{}, error) {
	log.Println("Create new catalog connection using catalog connector address: ", r.catalogConnectorAddress)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeOut)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, r.catalogConnectorAddress, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("Connection to External Catalog Connector failed: %v", err)
	}
	defer conn.Close()
	client := pb.NewDataCatalogServiceClient(conn)

	appID := in.GetAppId()

	// datasetID -> metadata of dataset in form of map
	datasetsMetadata := make(map[string]interface{})
	for _, datasetContext := range in.GetDatasets() {
		dataset := datasetContext.GetDataset()
		datasetID := dataset.GetDatasetId()

		if _, present := datasetsMetadata[datasetID]; !present {
			metadataMap, err := r.GetDatasetMetadata(&ctx, client, datasetID, appID)

			if err != nil {
				return nil, err
			}
			datasetsMetadata[datasetID] = metadataMap
		}
	}

	return datasetsMetadata, nil
}

func (r *CatalogReader) GetDatasetMetadata(ctx *context.Context, client pb.DataCatalogServiceClient, datasetID string, appID string) (map[string]interface{}, error) {
	objToSend := &pb.CatalogDatasetRequest{AppId: appID, DatasetId: datasetID}
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
