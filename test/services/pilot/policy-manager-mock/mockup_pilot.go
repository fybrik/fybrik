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

	"emperror.dev/errors"
	connectors "github.com/mesh-for-data/mesh-for-data/pkg/connectors/clients"
	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	setupLog = ctrl.Log.WithName("setup")
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
		datasetIDJson = "{\"ServerName\":\"mds1\",\"AssetGuid\":\"1e2a0403-1946-4e89-a10b-fd96eda5a5dc\"}"
	} else {
		datasetIDJson = "{\"catalog_id\":\"" + catalogID + "\",\"asset_id\":\"" + datasetID + "\"}"
	}

	applicationDetails := &pb.ApplicationDetails{Properties: map[string]string{"intent": "fraud-detection"}, ProcessingGeography: "US"}

	datasets := []*pb.DatasetContext{}
	datasets = append(datasets, createDatasetRead(datasetIDJson))
	datasets = append(datasets, createDatasetTransferFirst(datasetIDJson))
	// datasets = append(datasets, createDatasetTransferSecond(catalogID, datasetID))
	// datasets = append(datasets, createDatasetRead(catalogIDcos, datasetIDcos))
	// datasets = append(datasets, createDatasetRead(catalogIDDb2, datasetIDDb2))
	// datasets = append(datasets, createDatasetTransferFirst(catalogIDDb2, datasetIDDb2))
	// datasets = append(datasets, createDatasetTransferSecond(catalogIDDb2, datasetIDDb2))

	// applicationContext := &pb.ApplicationContext{PolicyManagerCredentials: policyManagerCredentials, AppInfo: applicationDetails, Datasets: datasets}
	applicationContext := &pb.ApplicationContext{AppInfo: applicationDetails, Datasets: datasets}
	log.Printf("Sending Application Context: ")
	appContextStr, _ := json.MarshalIndent(applicationContext, "", "\t")
	log.Print(string(appContextStr))
	log.Println("1***************************************************************")

	return applicationContext
}

// TODO: newPolicyManager is a duplicate of newPolicyManager from main.go

func newPolicyManager() (connectors.PolicyManager, error) {
	connectionTimeout := os.Getenv("CONNECTION_TIMEOUT")
	timeOutInSeconds, err := strconv.Atoi(connectionTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "Atoi conversion of CONNECTION_TIMEOUT failed")
	}

	mainPolicyManagerName := os.Getenv("MAIN_POLICY_MANAGER_NAME")
	mainPolicyManagerURL := os.Getenv("MAIN_POLICY_MANAGER_CONNECTOR_URL")
	policyManager, err := connectors.NewGrpcPolicyManager(
		mainPolicyManagerName, mainPolicyManagerURL, time.Duration(timeOutInSeconds)*time.Second)
	setupLog.Info("setting main policy manager", "Name", mainPolicyManagerName, "URL", mainPolicyManagerURL, "Timeout (sec)", timeOutInSeconds)
	if err != nil {
		return nil, err
	}

	useExtensionPolicyManager, err := strconv.ParseBool(os.Getenv("USE_EXTENSIONPOLICY_MANAGER"))
	if useExtensionPolicyManager && err == nil {
		extensionPolicyManagerName := os.Getenv("EXTENSIONS_POLICY_MANAGER_NAME")
		extensionPolicyManagerURL := os.Getenv("EXTENSIONS_POLICY_MANAGER_CONNECTOR_URL")
		extensionPolicyManager, err := connectors.NewGrpcPolicyManager(
			extensionPolicyManagerName, extensionPolicyManagerURL, time.Duration(timeOutInSeconds)*time.Second)
		setupLog.Info("setting extension policy manager", "Name", extensionPolicyManagerName, "URL", extensionPolicyManagerURL, "Timeout (sec)", timeOutInSeconds)
		if err != nil {
			return nil, err
		}

		policyManager = connectors.NewMultiPolicyManager(policyManager, extensionPolicyManager)
	}

	return policyManager, nil
}

func main() {
	applicationContext := constructInputParameters()

	policyManager, err := newPolicyManager()
	if err != nil {
		setupLog.Error(err, "unable to create policy manager facade")
		os.Exit(1)
	}

	r, err := policyManager.GetPoliciesDecisions(context.Background(), applicationContext)

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
	r, err = policyManager.GetPoliciesDecisions(context.Background(), applicationContext)

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
