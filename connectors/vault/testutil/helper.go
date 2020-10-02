// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"log"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
)

var EnvValues = map[string]string{
	"CONNECTION_TIMEOUT":    "120",
	"CATALOG_CONNECTOR_URL": "localhost:50084",
	"OPA_SERVER_URL":        "localhost:8282",
	"USER_VAULT_PATH":       "external",
}

func GetEnvironment() string {
	userVaultPath := EnvValues["USER_VAULT_PATH"]
	log.Printf("EnvVariables = %v\n", EnvValues)

	return userVaultPath
}

func createDatasetRead(datasetID string) *pb.DatasetContext {
	dataset := &pb.DatasetIdentifier{DatasetId: datasetID}
	operation := &pb.AccessOperation{Type: pb.AccessOperation_READ}
	datasetContext := &pb.DatasetContext{Dataset: dataset, Operation: operation}
	return datasetContext
}

func createDatasetTransferFirst(datasetID string) *pb.DatasetContext {
	dataset := &pb.DatasetIdentifier{DatasetId: datasetID}
	operation := &pb.AccessOperation{Type: pb.AccessOperation_COPY, Destination: "US"}
	datasetContext := &pb.DatasetContext{Dataset: dataset, Operation: operation}
	return datasetContext
}

func EnsureDeepEqualCredentials(t *testing.T, testedCredentials *pb.DatasetCredentials, expectedCredentials *pb.DatasetCredentials) {
	assert.True(t, proto.Equal(testedCredentials, expectedCredentials), "DatasetCredentials we got are not as expected. Expected: %v, Received: %v", expectedCredentials, testedCredentials)
}

func GetExpectedVaultCredentials(in *pb.DatasetCredentialsRequest) *pb.DatasetCredentials {
	return &pb.DatasetCredentials{DatasetId: "mock-datasetID",
		Credentials: "{\"credentials\":\"my_egeria_credentials_test\",\"dataset_id\":\"mock-datasetID\"}"}
}
