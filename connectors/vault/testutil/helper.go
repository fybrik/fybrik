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

func EnsureDeepEqualCredentials(t *testing.T, testedCredentials *pb.DatasetCredentials, expectedCredentials *pb.DatasetCredentials) {
	assert.True(t, proto.Equal(testedCredentials, expectedCredentials), "DatasetCredentials we got are not as expected. Expected: %v, Received: %v", expectedCredentials, testedCredentials)
}

func GetExpectedVaultCredentials(in *pb.DatasetCredentialsRequest) *pb.DatasetCredentials {
	credentials := &pb.Credentials{AccessKey: "dummy_access_key", SecretKey: "dummy_secret_key"}
	return &pb.DatasetCredentials{DatasetId: "mock-datasetID",
		Creds: credentials}
}
