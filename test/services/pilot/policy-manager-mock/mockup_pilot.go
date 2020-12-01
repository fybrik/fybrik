// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
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

func constructInputParameters() *pb.ApplicationContext {
	timeoutinsecs := getEnv("CONNECTION_TIMEOUT")
	timeoutinseconds, err := strconv.Atoi(timeoutinsecs)
	if err != nil {
		log.Printf("Atoi conversion of timeoutinseconds failed: %v", err)
		return nil
	}

	fmt.Println("timeoutinseconds in MockupPilot: ", timeoutinseconds)

	// Defining an applicationcontext
	// ownerID := "999"
	// credentialsStr := "{ ownerID: " + ownerID + "}"

	// example 0: local file
	// catalogID := "cc17803b-163a-43db-97e3-323a8519c78f"  //democatalog
	// datasetID := "6c49313f-1207-4995-a957-5cd49c4e57ac"

	// example 1: remote parquet
	// datasetIDcos := "10a9fba1-b049-40d9-bac9-1a608c1e4774" //small.parq
	// catalogIDcos := "591258ed-7461-47db-8eb6-1edf285c26cd" //EtyCatalog

	// example 2: remote db2
	// datasetIDDb2 := "2d1b5352-1fbf-439b-8bb0-c1967ac484b9" //Connection-Db2-NQD60833-SMALL
	// catalogIDDb2 := "1c080331-72da-4cea-8d06-5f075405cf17" //catalog-suri

	// example 3: remote csv,
	// datasetID := "79aaff22-cfbe-470a-86b6-8f5125781a5c"
	// catalogID := "1c080331-72da-4cea-8d06-5f075405cf17"

	// kafka
	catalogID := "87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f"
	datasetID := "01c6f0f0-9ffe-4ccc-ac07-409523755e72" // "988f7b32-2417-4b4f-b327-d4a63d110267" // "466b5d7c-38c5-438c-8298-5c7e00e40638"

	var datasetIDJson string
	if getEnv("CATALOG_PROVIDER_NAME") == "EGERIA" {
		// datasetIDJson = "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"24cd3ed9-4084-43b9-9e91-5fe1f4fbd6b7\"}"
		datasetIDJson = "{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"1e2a0403-1946-4e89-a10b-fd96eda5a5dc\"}"
	} else {
		datasetIDJson = "{\"catalog_id\":\"" + catalogID + "\",\"asset_id\":\"" + datasetID + "\"}"
	}

	applicationDetails := &pb.ApplicationDetails{Purpose: "fraud-detection", Role: "Security", ProcessingGeography: "US"}

	appID := "datauser1/notebook-with-kafka"
	// credentialsStr := "v1/m4d/user-creds/datauser1/notebook-with-kafka/WKC"
	// policyManagerCredentials := &pb.PolicyManagerCredentials{Credentials: credentialsStr}

	datasets := []*pb.DatasetContext{}
	datasets = append(datasets, createDatasetRead(datasetIDJson))
	datasets = append(datasets, createDatasetTransferFirst(datasetIDJson))
	// datasets = append(datasets, createDatasetTransferSecond(catalogID, datasetID))
	// datasets = append(datasets, createDatasetRead(catalogIDcos, datasetIDcos))
	// datasets = append(datasets, createDatasetRead(catalogIDDb2, datasetIDDb2))
	// datasets = append(datasets, createDatasetTransferFirst(catalogIDDb2, datasetIDDb2))
	// datasets = append(datasets, createDatasetTransferSecond(catalogIDDb2, datasetIDDb2))

	// applicationContext := &pb.ApplicationContext{PolicyManagerCredentials: policyManagerCredentials, AppInfo: applicationDetails, Datasets: datasets}
	applicationContext := &pb.ApplicationContext{AppId: appID, AppInfo: applicationDetails, Datasets: datasets}
	log.Printf("Sending Application Context: ")
	appContextStr, _ := json.MarshalIndent(applicationContext, "", "\t")
	log.Print(string(appContextStr))
	log.Println("1***************************************************************")

	return applicationContext
}
func main() {
	applicationContext := constructInputParameters()

	policyCompiler := pc.NewPolicyCompiler()
	r, err := policyCompiler.GetPoliciesDecisions(applicationContext)

	if err != nil {
		errStatus, _ := status.FromError(err)
		fmt.Println("*********************************in error in  MockupPilot *****************************")
		fmt.Println("Message: ", errStatus.Message())
		fmt.Println("Code: ", errStatus.Code())

		// take specific action based on specific error?
		if codes.InvalidArgument == errStatus.Code() {
			fmt.Println("InvalidArgument in mockup pilot")
			return
		}
	} else {
		log.Printf("Response received from Policy Compiler below:")
		s, _ := json.MarshalIndent(r, "", "    ")
		log.Print(string(s))
		log.Println("2***************************************************************")
	}

	fmt.Println("*********************************invoking new request *****************************")
	r, err = policyCompiler.GetPoliciesDecisions(applicationContext)

	if err != nil {
		errStatus, _ := status.FromError(err)
		fmt.Println("*********************************in error in  MockupPilot for 2nd request *****************************")
		fmt.Println("Message: ", errStatus.Message())
		fmt.Println("Code: ", errStatus.Code())

		// take specific action based on specific error?
		if codes.InvalidArgument == errStatus.Code() {
			fmt.Println("InvalidArgument in mockup pilot for 2nd request")
			return
		}
	} else {
		log.Printf("Response received from Policy Compiler below for 2nd request:")
		s, _ := json.MarshalIndent(r, "", "    ")
		log.Print(string(s))
	}
}

func createDatasetRead(datasetIDJson string) *pb.DatasetContext {
	dataset := &pb.DatasetIdentifier{DatasetId: datasetIDJson}
	operation := &pb.AccessOperation{Type: pb.AccessOperation_READ}
	datasetContext := &pb.DatasetContext{Dataset: dataset, Operation: operation}
	return datasetContext
}

func createDatasetTransferFirst(datasetIDJson string) *pb.DatasetContext {
	dataset := &pb.DatasetIdentifier{DatasetId: datasetIDJson}
	operation := &pb.AccessOperation{Type: pb.AccessOperation_COPY, Destination: "US"}
	datasetContext := &pb.DatasetContext{Dataset: dataset, Operation: operation}
	return datasetContext
}

// func createDatasetTransferSecond(catalogID, datasetID string) *pb.DatasetContext {
// 	dataset := &pb.DatasetIdentifier{CatalogId: catalogID, DatasetId: datasetID}
// 	operation := &pb.AccessOperation{Type: pb.AccessOperation_COPY, Destination: "European Union"}
// 	datasetContext := &pb.DatasetContext{Dataset: dataset, Operation: operation}
// 	return datasetContext
// }
