// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Env Variable %v not defined", key)
	}
	return value
}

func GetMetadata(datasetID string) error {
	catalogConnectorURL := getEnv("CATALOG_CONNECTOR_URL")
	catalogProviderName := getEnv("CATALOG_PROVIDER_NAME")

	timeoutInSecs := getEnv("CONNECTION_TIMEOUT")
	timeoutInSeconds, err := strconv.Atoi(timeoutInSecs)
	if err != nil {
		log.Printf("Atoi conversion of timeoutinseconds failed: %v", err)
		return errors.Wrap(err, "Atoi conversion of timeoutinseconds failed")
	}

	fmt.Println("timeoutInSeconds: ", timeoutInSeconds)
	fmt.Println("catalogConnectorURL: ", catalogConnectorURL)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutInSeconds)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, catalogConnectorURL, grpc.WithInsecure())

	if err != nil {
		log.Printf("Connection to "+catalogProviderName+" Catalog Connector failed: %v", err)
		errStatus, _ := status.FromError(err)
		fmt.Println(errStatus.Message())
		fmt.Println(errStatus.Code())
		return errors.Wrap(err, "Connection to "+catalogProviderName+" Catalog Connector failed")
	}
	defer conn.Close()

	c := pb.NewDataCatalogServiceClient(conn)

	// catalogCredentials := "v1/m4d/user-creds/datauser1/notebook-with-kafka/WKC"
	appID := "datauser1/notebook-with-kafka"
	objToSend := &pb.CatalogDatasetRequest{AppId: appID, DatasetId: datasetID}

	log.Printf("Sending CatalogDatasetRequest: ")
	catDataReqStr, _ := json.MarshalIndent(objToSend, "", "\t")
	log.Print(string(catDataReqStr))
	log.Println("1***************************************************************")

	log.Printf("Sending request to " + catalogProviderName + " Catalog Connector Server")
	r, err := c.GetDatasetInfo(ctx, objToSend)

	// updated for better exception handling using standard GRPC codes
	if err != nil {
		log.Printf("Error sending data to %s Catalog Connector: %v", catalogProviderName, err)
		errStatus, _ := status.FromError(err)
		fmt.Println("Message:", errStatus.Message())
		// lets print the error code which is `INVALID_ARGUMENT`
		fmt.Println("Code:", errStatus.Code())
		return errors.Wrap(err, "Error sending data to Catalog Connector")
	}

	fmt.Println("***************************************************************")
	log.Printf("Received Response for GetDatasetInfo with  datasetID: %s\n", r.GetDatasetId())
	fmt.Println("***************************************************************")
	log.Printf("Response received from %s is given below:", catalogProviderName)
	s, _ := json.MarshalIndent(r, "", "\t")
	fmt.Print(string(s))
	fmt.Println("***************************************************************")
	return nil
}

func GetCredentials(datasetID string) error {
	credentialsConnectorURL := getEnv("CREDENTIALS_CONNECTOR_URL")
	credentialsProviderName := getEnv("CREDENTIALS_PROVIDER_NAME")

	timeoutInSecs := getEnv("CONNECTION_TIMEOUT")
	timeoutInSeconds, err := strconv.Atoi(timeoutInSecs)

	if err != nil {
		log.Printf("Atoi conversion of timeoutinseconds failed: %v", err)
		return errors.Wrap(err, "Atoi conversion of timeoutinseconds failed in GetCredentials")
	}

	fmt.Println("timeoutInSeconds: ", timeoutInSeconds)
	fmt.Println("credentialsConnectorURL: ", credentialsConnectorURL)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutInSeconds)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, credentialsConnectorURL, grpc.WithInsecure())

	if err != nil {
		log.Printf("Connection to "+credentialsProviderName+" Credentials Connector failed: %v", err)
		errStatus, _ := status.FromError(err)
		fmt.Println(errStatus.Message())

		fmt.Println(errStatus.Code())
		return errors.Wrap(err, "Connection to Credentials Connector failed")
	}
	defer conn.Close()

	c1 := pb.NewDataCredentialServiceClient(conn)

	// userCredentials := "v1/m4d/user-creds/datauser1/notebook-with-kafka/WKC"
	appID := "datauser1/notebook-with-kafka"
	objToSendForCredential := &pb.DatasetCredentialsRequest{AppId: appID, DatasetId: datasetID}

	log.Printf("Sending DatasetCredentialsRequest: ")
	dataCredReqStr, _ := json.MarshalIndent(objToSendForCredential, "", "\t")
	log.Print(string(dataCredReqStr))
	log.Println("1***************************************************************")

	log.Println("Sending request to " + credentialsProviderName + " Connector Server")
	responseCredential, errCredential := c1.GetCredentialsInfo(ctx, objToSendForCredential)

	// updated for better exception handling using standard GRPC codes
	if errCredential != nil {
		log.Printf("Error sending data to "+credentialsProviderName+" Credentials Connector: %v", errCredential)
		errCredentialStatus, _ := status.FromError(errCredential)
		log.Println("Message:", errCredentialStatus.Message())
		// lets print the error code which is `INVALID_ARGUMENT`
		log.Println("Code:", errCredentialStatus.Code())
		return errors.Wrap(err, "Error sending data to Credentials Connector in GetCredentials")
	}

	log.Println("***************************************************************")
	log.Printf("Received Response for GetCredentialsInfo with datasetID: %s\n", responseCredential.GetDatasetId())
	log.Println("***************************************************************")
	log.Printf("Response received from %s is given below:", credentialsProviderName)
	sCredential, _ := json.MarshalIndent(responseCredential, "", "\t")
	log.Print(string(sCredential))
	log.Println("***************************************************************")
	return nil
}

func main() {
	// example 1: remote parquet
	// datasetID := "10a9fba1-b049-40d9-bac9-1a608c1e4774"
	// catalogID := "591258ed-7461-47db-8eb6-1edf285c26cd"

	// example 2: remote db2
	// datasetID := "2d1b5352-1fbf-439b-8bb0-c1967ac484b9"
	// catalogID := "1c080331-72da-4cea-8d06-5f075405cf17"

	// example 3: remote csv,
	// datasetID := "79aaff22-cfbe-470a-86b6-8f5125781a5c";
	// catalogID := "1c080331-72da-4cea-8d06-5f075405cf17";

	// example 4: local csv
	// datasetID := "cc17803b-163a-43db-97e3-323a8519c78f"
	// dataSetID := "6c49313f-1207-4995-a957-5cd49c4e57ac"

	// kafka
	catalogID := "87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f"
	datasetID := "01c6f0f0-9ffe-4ccc-ac07-409523755e72" // "466b5d7c-38c5-438c-8298-5c7e00e40638"

	var datasetIDJson string
	if getEnv("CATALOG_PROVIDER_NAME") == "EGERIA" {
		// datasetIDJson = "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"24cd3ed9-4084-43b9-9e91-5fe1f4fbd6b7\"}"
		datasetIDJson = "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"1e2a0403-1946-4e89-a10b-fd96eda5a5dc\"}"
	} else {
		datasetIDJson = "{\"catalog_id\":\"" + catalogID + "\",\"asset_id\":\"" + datasetID + "\"}"
	}

	err := GetMetadata(datasetIDJson)
	if err != nil {
		fmt.Printf("Error in GetCredentials:\n %v\n\n", err)
		fmt.Printf("Error in GetCredentials Details:\n%+v\n\n", err)
		// errors.Cause() provides access to original error.
		fmt.Printf("Error in GetCredentials Cause: %v\n", errors.Cause(err))
		fmt.Printf("Error in GetCredentials Extended Cause:\n%+v\n", errors.Cause(err))
	}

	err = GetCredentials(datasetIDJson)
	if err != nil {
		fmt.Printf("Error in GetCredentials: \n %v\n\n", err)
		fmt.Printf("Error in GetCredentials Details: \n%+v\n\n", err)
		// errors.Cause() provides access to original error.
		fmt.Printf("Error in GetCredentials Cause: %v\n", errors.Cause(err))
		fmt.Printf("Error in GetCredentials Details Extended Cause:\n%+v\n", errors.Cause(err))
	}
}
