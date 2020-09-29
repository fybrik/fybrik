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

	vltutils "github.com/ibm/the-mesh-for-data/connectors/vault/vault_utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Env Variable %v not defined", key)
	}
	return value
}

func GetMetadata(datasetID string) {
	catalogConnectorURL := getEnv("CATALOG_CONNECTOR_URL")
	catalogProviderName := getEnv("CATALOG_PROVIDER_NAME")

	timeoutInSecs := getEnv("CONNECTION_TIMEOUT")
	timeoutInSeconds, err := strconv.Atoi(timeoutInSecs)
	if err != nil {
		log.Printf("Atoi conversion of timeoutinseconds failed: %v", err)
		return
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
		return
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
		// Want its int version for some reason?
		// you shouldn't actullay do this, but if you need for debugging,
		// you can do `int(status_code)` which will give you `3`
		//
		// Want to take specific action based on specific error?
		if codes.InvalidArgument == errStatus.Code() {
			// do your stuff here
			log.Fatal()
		}
	} else {
		fmt.Println("***************************************************************")
		log.Printf("Received Response for GetDatasetInfo with  datasetID: %s\n", r.GetDatasetId())
		fmt.Println("***************************************************************")
		log.Printf("Response received from %s is given below:", catalogProviderName)
		s, _ := json.MarshalIndent(r, "", "\t")
		fmt.Print(string(s))
		fmt.Println("***************************************************************")
	}
}

func GetCredentials(datasetID string) {
	credentialsConnectorURL := getEnv("CREDENTIALS_CONNECTOR_URL")
	credentialsProviderName := getEnv("CREDENTIALS_PROVIDER_NAME")

	timeoutInSecs := getEnv("CONNECTION_TIMEOUT")
	timeoutInSeconds, err := strconv.Atoi(timeoutInSecs)

	if err != nil {
		log.Printf("Atoi conversion of timeoutinseconds failed: %v", err)
		return
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
		return
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
		// Want its int version for some reason?
		// you shouldn't actullay do this, but if you need for debugging,
		// you can do `int(status_code)` which will give you `3`
		//
		// Want to take specific action based on specific error?
		if codes.InvalidArgument == errCredentialStatus.Code() {
			// do your stuff here
			log.Fatal()
		}
	} else {
		log.Println("***************************************************************")
		log.Printf("Received Response for GetCredentialsInfo with datasetID: %s\n", responseCredential.GetDatasetId())
		log.Println("***************************************************************")
		log.Printf("Response received from %s is given below:", credentialsProviderName)
		sCredential, _ := json.MarshalIndent(responseCredential, "", "\t")
		log.Print(string(sCredential))
		log.Println("***************************************************************")
	}
}

func ConfigureVault(innerVaultPath string, credentials string) error {
	vaultAddress := vltutils.GetEnv(vltutils.VaultAddressKey)
	timeOutInSecs := vltutils.GetEnvWithDefault(vltutils.VaultTimeoutKey, vltutils.DefaultTimeout)
	timeOutSecs, err := strconv.Atoi(timeOutInSecs)
	port := vltutils.GetEnvWithDefault(vltutils.VaultConnectorPortKey, vltutils.DefaultPort)

	log.Printf("Vault address env variable in %s: %s\n", vltutils.VaultAddressKey, vaultAddress)
	log.Printf("VaultConnectorPort env variable in %s: %s\n", vltutils.VaultConnectorPortKey, port)
	log.Printf("TimeOut used %d\n", timeOutSecs)
	log.Printf("Secret Token env variable in %s: %s\n", vltutils.VaultSecretKey, vltutils.GetEnv(vltutils.VaultSecretKey))

	var vault vltutils.VaultConnection
	vault = vltutils.CreateVaultConnection()
	log.Println("Vault connection successfully initiated.")

	credentialsMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(credentials), &credentialsMap); err != nil {
		log.Println("err in json.Unmarshal")
		errStatus, _ := status.FromError(err)
		log.Println("Message:", errStatus.Message())
		log.Println("Code:", errStatus.Code())
		return err
	}

	vaultPath := vltutils.GetEnv(vltutils.VaultPathKey) + "/" + innerVaultPath
	if _, err := vault.AddToVault(vaultPath, credentialsMap); err != nil {
		log.Println("err in utils.AddToVault")
		errStatus, _ := status.FromError(err)
		log.Println("Message:", errStatus.Message())
		log.Println("Code:", errStatus.Code())
		return err
	}

	var retrievedValue string
	if retrievedValue, err = vault.GetFromVault(innerVaultPath); err != nil {
		log.Println("err in utils.GetFromVault")
		errStatus, _ := status.FromError(err)
		log.Println("Message:", errStatus.Message())
		log.Println("Code:", errStatus.Code())
		return err
	}
	log.Println("retrievedValue from vault:", retrievedValue)
	return nil

}

func main() {
	//example 1: remote parquet
	// datasetID := "10a9fba1-b049-40d9-bac9-1a608c1e4774"
	// catalogID := "591258ed-7461-47db-8eb6-1edf285c26cd"

	//example 2: remote db2
	// datasetID := "2d1b5352-1fbf-439b-8bb0-c1967ac484b9"
	// catalogID := "1c080331-72da-4cea-8d06-5f075405cf17"

	//example 3: remote csv,
	//datasetID := "79aaff22-cfbe-470a-86b6-8f5125781a5c";
	//catalogID := "1c080331-72da-4cea-8d06-5f075405cf17";

	//example 4: local csv
	// datasetID := "cc17803b-163a-43db-97e3-323a8519c78f"
	// dataSetID := "6c49313f-1207-4995-a957-5cd49c4e57ac"

	//kafka
	catalogID := "87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f"
	datasetID := "01c6f0f0-9ffe-4ccc-ac07-409523755e72" //"466b5d7c-38c5-438c-8298-5c7e00e40638"

	var datasetIDJson string
	if getEnv("CATALOG_PROVIDER_NAME") == "EGERIA" {
		// datasetIDJson = "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"4098e18e-bd53-4fd0-8ff8-e1c8e9fc42da\"}"
		datasetIDJson = "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"91aec690-bf78-4172-9ef2-cd0abd74b4b1\"}"
	} else {
		datasetIDJson = "{\"catalog_id\":\"" + catalogID + "\",\"asset_id\":\"" + datasetID + "\"}"
	}

	ConfigureVault(datasetIDJson, "{\"credentials\": \"my_egeria_credentials\"}")
	GetMetadata(datasetIDJson)
	GetCredentials(datasetIDJson)
}
