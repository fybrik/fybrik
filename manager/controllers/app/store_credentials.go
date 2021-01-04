// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"

	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
)

// RegisterCredentials stores credentials in vault
// Credentials are stored as a string (JSON). Using JSON allows providing different data stores with different types of credentials
// Credentials are received from an external credential manager and are stored by Pilot as-is.
func (r *M4DApplicationReconciler) RegisterCredentials(req *modules.DataInfo, credentials *pb.Credentials) error {
	jsonStr, err := json.Marshal(credentials)
	if err != nil {
		return err
	}
	credentialsMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonStr, &credentialsMap); err != nil {
		return err
	}
	if _, err := utils.AddToVault(req.AssetID, credentialsMap, r.VaultClient); err != nil {
		return err
	}
	req.Credentials = utils.GetDatasetVaultPath(req.AssetID)
	return nil
}
