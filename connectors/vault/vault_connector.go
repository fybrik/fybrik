// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	vltutils "github.com/ibm/the-mesh-for-data/connectors/vault/vault_utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

var vault vltutils.VaultConnection

const defaultTimeout = "180"
const defaultPort = "50083" //synched with vault_connector.yaml

type server struct {
	pb.UnimplementedDataCredentialServiceServer
}

func (s *server) GetCredentialsInfo(ctx context.Context, in *pb.DatasetCredentialsRequest) (*pb.DatasetCredentials, error) {
	log.Println("Vault connector: GetCredentialsInfo")
	log.Printf("Vault connector: GetCredentialsInfo, in = %v\n", in)

	// secretaddr := url.QueryEscape(in.DatasetId)
	secretaddr := strings.ReplaceAll(in.DatasetId, " ", "")

	readCredentials, err := vault.GetFromVault(secretaddr)
	if err != nil {
		log.Printf("Error in vaultConnector, got error from the vault: %v\n", err.Error())
		return nil, fmt.Errorf("error in retrieving the secret from vault(key = %s): %v", secretaddr, err)
	}
	log.Println("Read credentials from vault: " + readCredentials)

	return &pb.DatasetCredentials{DatasetId: in.DatasetId, Credentials: readCredentials}, nil
}

func main() {
	vaultAddress := vltutils.GetEnv(vltutils.VaultAddressKey)
	timeOutInSecs := vltutils.GetEnvWithDefault(vltutils.VaultTimeoutKey, vltutils.DefaultTimeout)
	timeOutSecs, err := strconv.Atoi(timeOutInSecs)
	port := vltutils.GetEnvWithDefault(vltutils.VaultConnectorPortKey, vltutils.DefaultPort)

	log.Printf("Vault address env variable in %s: %s\n", vltutils.VaultAddressKey, vaultAddress)
	log.Printf("VaultConnectorPort env variable in %s: %s\n", vltutils.VaultConnectorPortKey, port)
	log.Printf("TimeOut used %d\n", timeOutSecs)
	log.Printf("Secret Token env variable in %s: %s\n", vltutils.VaultSecretKey, vltutils.GetEnv(vltutils.VaultSecretKey))

	vault = vltutils.CreateVaultConnection()
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
