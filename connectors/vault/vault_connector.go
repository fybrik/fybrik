// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	vaultutils "github.com/ibm/the-mesh-for-data/connectors/vault/vault_utils"
	vaultconnbl "github.com/ibm/the-mesh-for-data/connectors/vault/vaultconn-bl"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

const defaultTimeout = "180"
const defaultPort = "50083" //synched with vault_connector.yaml

type server struct {
	pb.UnimplementedDataCredentialServiceServer
	vaultConnector *vaultconnbl.Server
}

var vault vaultutils.VaultConnection

func (s *server) GetCredentialsInfo(ctx context.Context, in *pb.DatasetCredentialsRequest) (*pb.DatasetCredentials, error) {
	log.Println("Received ApplicationContext")
	log.Println(in)

	vaultPathKey := vaultutils.GetEnv(vaultutils.VaultPathKey)
	eval, err := s.vaultConnector.GetCredentialsInfo(in, vault, vaultPathKey)
	if err != nil {
		log.Printf("Error in vaultConnector, got error from the vault: %v\n", err.Error())
		return nil, fmt.Errorf("error in retrieving the secret from vault: %v", err)
	}
	log.Println("Received evaluation from vaultConnector.GetCredentialsInfo ")
	return eval, err
}

func main() {
	vaultAddress := vaultutils.GetEnv(vaultutils.VaultAddressKey)
	timeOutInSecs := vaultutils.GetEnvWithDefault(vaultutils.VaultTimeoutKey, defaultTimeout)
	timeOutSecs, err := strconv.Atoi(timeOutInSecs)
	port := vaultutils.GetEnvWithDefault(vaultutils.VaultConnectorPortKey, defaultPort)

	log.Printf("Vault address env variable in %s: %s\n", vaultutils.VaultAddressKey, vaultAddress)
	log.Printf("VaultConnectorPort env variable in %s: %s\n", vaultutils.VaultConnectorPortKey, port)
	log.Printf("TimeOut used %d\n", timeOutSecs)
	log.Printf("Secret Token env variable is accessed from environment %s\n", vaultutils.VaultSecretKey)

	vault = vaultutils.CreateVaultConnection()
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
