// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"

	. "github.com/ibm/the-mesh-for-data/connectors/vault/vault_utils"
	vaultbl "github.com/ibm/the-mesh-for-data/connectors/vault/vaultconn-bl"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

const defaultTimeout = "180"
const defaultPort = "50083" //synched with vault_connector.yaml

type server struct {
	pb.UnimplementedDataCredentialServiceServer
	vaultConnector *vaultbl.Server
}

var vault VaultConnection

func (s *server) GetCredentialsInfo(ctx context.Context, in *pb.DatasetCredentialsRequest) (*pb.DatasetCredentials, error) {
	log.Println("Received ApplicationContext")
	log.Println(in)

	vaultPathKey := GetEnv(VaultPathKey)
	eval, err := s.vaultConnector.GetCredentialsInfo(in, vault, vaultPathKey)
	if err != nil {
		log.Printf("Error in vaultConnector, got error from the vault: %v\n", err.Error())
		return nil, fmt.Errorf("error in retrieving the secret from vault: %v", err)
	}
	jsonOutput, _ := json.MarshalIndent(eval, "", "\t")
	log.Println("Received evaluation : " + string(jsonOutput))
	log.Println("err:", err)

	return eval, err

}

func main() {
	vaultAddress := GetEnv(VaultAddressKey)
	timeOutInSecs := GetEnvWithDefault(VaultTimeoutKey, defaultTimeout)
	timeOutSecs, err := strconv.Atoi(timeOutInSecs)
	port := GetEnvWithDefault(VaultConnectorPortKey, defaultPort)

	log.Printf("Vault address env variable in %s: %s\n", VaultAddressKey, vaultAddress)
	log.Printf("VaultConnectorPort env variable in %s: %s\n", VaultConnectorPortKey, port)
	log.Printf("TimeOut used %d\n", timeOutSecs)
	log.Printf("Secret Token env variable in %s: %s\n", VaultSecretKey, GetEnv(VaultSecretKey))

	vault = CreateVaultConnection()
	log.Println("Vault connection successfully initiated.")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error in listening: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDataCredentialServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error in service: %v", err)
	}
}
