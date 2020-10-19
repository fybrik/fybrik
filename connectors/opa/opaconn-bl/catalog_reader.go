// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package opaconnbl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//Reader from catalog, single instance for the connector, not dependent on the request
type CatalogReader struct {
	catalogConnectorAddress string
	timeOut                 int
}

func NewCatalogReader(address string, timeOut int) *CatalogReader {
	return &CatalogReader{catalogConnectorAddress: address, timeOut: timeOut}
}

//return map  datasetID -> metadata of dataset in form of map
func (r *CatalogReader) GetDatasetsMetadataFromCatalog(in *pb.ApplicationContext) (map[string]interface{}, error) {
	log.Println("Create new catalog connection using catalog connector address: ", r.catalogConnectorAddress)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeOut)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, r.catalogConnectorAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("Connection to External Catalog Connector failed: %v", err)
		errStatus, _ := status.FromError(err)
		log.Println(errStatus.Message())
		log.Println(errStatus.Code())
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

func (c *CatalogReader) GetDatasetMetadata(ctx *context.Context, client pb.DataCatalogServiceClient, datasetID string, appID string) (map[string]interface{}, error) {
	objToSend := &pb.CatalogDatasetRequest{AppId: appID, DatasetId: datasetID}
	log.Printf("Sending request to External Catalog Connector: datasetID = %s", datasetID)
	r, err := client.GetDatasetInfo(*ctx, objToSend)
	if err != nil {
		log.Printf("error sending data to External Catalog Connector (datasetID = %s): %v", datasetID, err)
		errStatus, _ := status.FromError(err)
		log.Println("Message:", errStatus.Message())
		log.Println("Code:", errStatus.Code())
		if codes.InvalidArgument == errStatus.Code() {
			log.Println("Invalid argument error : " + err.Error())
		}
		return nil, fmt.Errorf("error sending data to External Catalog Connector (datasetID = %s): %v", datasetID, err)
	}
	log.Println("***************************************************************")
	log.Printf("Received Response from External Catalog Connector for  dataSetID: %s\n", datasetID)
	log.Println("***************************************************************")
	log.Printf("Response received from External Catalog Connector is given below:")
	responseBytes, errJSON := json.MarshalIndent(r, "", "\t")
	if errJSON != nil {
		log.Printf("error Marshalling Catalog External Connector Response: %v", errJSON)
		return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
	}
	log.Print(string(responseBytes))
	log.Println("***************************************************************")
	metadataMap := make(map[string]interface{})
	err = json.Unmarshal(responseBytes, &metadataMap)
	if err != nil {
		log.Printf("error in unmarshalling responseBytes: %v", err)
		return nil, fmt.Errorf("error in unmarshalling responseBytes (datasetID = %s): %v", datasetID, err)
	}

	return metadataMap, nil
}
