// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vaultconnbl

import (
	"fmt"
	"log"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	vaultutils "github.com/ibm/the-mesh-for-data/connectors/vault/vault_utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
)

type Server struct {
}

func (s *Server) GetCredentialsInfo(in *pb.DatasetCredentialsRequest, vault vaultutils.VaultConnection, vaultPathKey string) (*pb.DatasetCredentials, error) {
	log.Printf("Vault connector: GetCredentialsInfo, in = %v\n", in)

	secretaddr := strings.ReplaceAll(in.DatasetId, " ", "")

	readCredentials, err := vault.GetFromVault(vaultPathKey, secretaddr)
	if err != nil {
		log.Printf("Error in vaultConnector, got error from the vault: %v\n", err.Error())
		return nil, fmt.Errorf("error in retrieving the secret from vault(key = %s): %v", secretaddr, err)
	}

	credentials := pb.Credentials{}
	err = jsonpb.UnmarshalString(readCredentials, &credentials)
	log.Println("readCredentials: unmarshalled into Credentials object ")

	if err != nil {
		log.Println("GetCredentialsInfo: Error - ", err)
		return nil, fmt.Errorf("error in UnmarshalString from readCredentials %s. Error is  %v", readCredentials, err)
	}
	dscredentials := &pb.DatasetCredentials{DatasetId: in.DatasetId, Creds: &credentials}
	log.Println("GetCredentialsInfo: Sending populated dscredentials object from vault connector ")
	return dscredentials, nil
}
