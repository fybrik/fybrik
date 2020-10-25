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
	//appID := "datauser1/notebook-with-kafka"
	appID := getEnv("APPID")
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
	//appID := "datauser1/notebook-with-kafka"
	appID := getEnv("APPID")
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

func ConfigureVault(vaultAddress string, outerVaultPath string, innerVaultPath string, credentials string) error {
	log.Printf("outerVaultPath used %s\n", outerVaultPath)
	log.Printf("innerVaultPath used %s\n", innerVaultPath)
	log.Printf("credentials used %s\n", credentials)

	timeOutInSecs := vltutils.GetEnvWithDefault(vltutils.VaultTimeoutKey, vltutils.DefaultTimeout)
	timeOutSecs, _ := strconv.Atoi(timeOutInSecs)
	port := vltutils.GetEnvWithDefault(vltutils.VaultConnectorPortKey, vltutils.DefaultPort)

	log.Printf("Vault address env variable in ConfigureVault: %s\n", vaultAddress)
	log.Printf("VaultConnectorPort env variable in %s: %s\n", vltutils.VaultConnectorPortKey, port)
	log.Printf("TimeOut used %d\n", timeOutSecs)
	log.Printf("Secret Token env variable in %s: %s\n", vltutils.VaultSecretKey, vltutils.GetEnv(vltutils.VaultSecretKey))

	vault := vltutils.CreateVaultConnection2(vaultAddress)
	log.Println("Vault connection successfully initiated.")

	credentialsMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(credentials), &credentialsMap); err != nil {
		log.Println("err in json.Unmarshal")
		errStatus, _ := status.FromError(err)
		log.Println("Message:", errStatus.Message())
		log.Println("Code:", errStatus.Code())
		return err
	}

	//vaultPath := vltutils.GetEnv(vltutils.VaultPathKey) + "/" + innerVaultPath
	vaultPath := outerVaultPath + "/" + innerVaultPath
	if _, err := vault.AddToVault(vaultPath, credentialsMap); err != nil {
		log.Println("err in utils.AddToVault")
		errStatus, _ := status.FromError(err)
		log.Println("Message:", errStatus.Message())
		log.Println("Code:", errStatus.Code())
		return err
	}

	var retrievedValue string
	var err error
	if retrievedValue, err =
		vault.GetFromVault2(outerVaultPath, innerVaultPath); err != nil {
		log.Println("err in utils.GetFromVault2")
		errStatus, _ := status.FromError(err)
		log.Println("Message:", errStatus.Message())
		log.Println("Code:", errStatus.Code())
		return err
	}
	log.Println("retrievedValue from vault:", retrievedValue)
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

	//kafka
	// catalogID := "87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f"
	// datasetID := "01c6f0f0-9ffe-4ccc-ac07-409523755e72" //"466b5d7c-38c5-438c-8298-5c7e00e40638"

	//cp4d 3 - ramasuri-catalog
	//catalogID := "8027121c-6da6-4093-9178-38f2062f5210"
	//datasetID := "c8367c56-c1ce-419e-996f-86b0c27d348e"
	//datasetID := "e7b19aba-a58c-4710-b7c8-077328d013cb"

	//cp4d 3 - Ritwik-catalog-testing-for-ING-catalog
	// catalogID := "fd0723d9-d604-4099-a2a6-fd7f2afe6dfa"
	// datasetID := "90e2f651-21b5-4f47-ac64-43b60f04f710"

	//cp4d 3 - Ritwik-catalog-testing-for-ING-v2
	// catalogID := "6a6acae2-026d-4342-acf7-9be4048aa0d3"
	// datasetID := "4af958ed-09a9-4ba6-8d02-c10efc1566a4"

	//cp4d 3 - New-Catalog-15-Oct-Demo
	catalogID := "fb4e5fc8-266c-4cf7-a29f-33e3c9d5d091"
	datasetID := "05df45f2-5927-4534-bdff-0d4b923baeea"

	var datasetIDJson string
	if getEnv("CATALOG_PROVIDER_NAME") == "EGERIA" {
		// datasetIDJson = "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"4098e18e-bd53-4fd0-8ff8-e1c8e9fc42da\"}"
		datasetIDJson = "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"f710567c-0f71-4296-b99e-cf22dc258a9f\"}"

		err := ConfigureVault(vltutils.GetEnv("USER_VAULT_ADDRESS"),
			vltutils.GetEnv("USER_VAULT_PATH"),
			datasetIDJson,
			"{\"credentials\": \"my_credentials\"}")
		if err != nil {
			log.Println("Error in ConfigureVault in mockup catalog1 ! ")
			return
		}
	} else {
		datasetIDJson = "{\"catalog_id\":\"" + catalogID + "\",\"asset_id\":\"" + datasetID + "\"}"

		// store in vault only in case we are using WKC
		wkcUserName := vltutils.GetEnv("CP4D_USERNAME_TO_BE_STORED_IN_VAULT")
		wkcPassword := vltutils.GetEnv("CP4D_PASSWORD_TO_BE_STORED_IN_VAULT")
		wkcOwnerID := vltutils.GetEnv("CP4D_OWNERID_TO_BE_STORED_IN_VAULT")
		appID := vltutils.GetEnv("APPID") + "/" + vltutils.GetEnv("CATALOG_PROVIDER_NAME")

		err := ConfigureVault(vltutils.GetEnv("VAULT_ADDRESS"),
			vltutils.GetEnv("VAULT_USER_HOME"),
			appID,
			"{\"username\":\""+wkcUserName+"\",\"password\":\""+wkcPassword+"\",\"ownerId\":\""+wkcOwnerID+"\"}")
		if err != nil {
			log.Println("Error in ConfigureVault in mockup catalog2 ! ")
			return
		}
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
