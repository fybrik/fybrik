// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vaultconnbl

import (
	"fmt"
	"log"
	"strings"

	vaultutils "github.com/ibm/the-mesh-for-data/connectors/vault/vault_utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
)

type Server struct {
}

func (s *Server) GetCredentialsInfo(in *pb.DatasetCredentialsRequest, vault vaultutils.VaultConnection, vaultPathKey string) (*pb.DatasetCredentials, error) {
	log.Println("Vault connector: GetCredentialsInfo")
	log.Printf("Vault connector: GetCredentialsInfo, in = %v\n", in)

	// secretaddr := url.QueryEscape(in.DatasetId)
	secretaddr := strings.ReplaceAll(in.DatasetId, " ", "")

	readCredentials, err := vault.GetFromVault(vaultPathKey, secretaddr)
	if err != nil {
		log.Printf("Error in vaultConnector, got error from the vault: %v\n", err.Error())
		return nil, fmt.Errorf("error in retrieving the secret from vault(key = %s): %v", secretaddr, err)
	}
	log.Println("Read credentials from vault: " + readCredentials)

	return &pb.DatasetCredentials{DatasetId: in.DatasetId, Credentials: readCredentials}, nil
}
