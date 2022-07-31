// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/rs/zerolog/log"

	fapp "fybrik.io/fybrik/manager/apis/app/v1"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/vault"
)

// RegisterAsset registers a new asset in the specified catalog
// Input arguments:
// - assetID: DataSetID as it appears in fybrik-application
// - catalogID: the destination catalog identifier
// - info: connection and credential details
// Returns:
// - an error if happened
// - the new asset identifier
func (r *FybrikApplicationReconciler) RegisterAsset(assetID string, catalogID string,
	info *fapp.DatasetDetails, input *fapp.FybrikApplication) (string, error) {
	r.Log.Trace().Msg("RegisterAsset")
	details := datacatalog.ResourceDetails{}
	if info.Details != nil {
		details.Connection = info.Details.Connection
		details.DataFormat = info.Details.Format
	}

	var resourceMetadata datacatalog.ResourceMetadata
	if info.ResourceMetadata != nil {
		resourceMetadata = *info.ResourceMetadata.DeepCopy()
	} else {
		resourceMetadata.Name = assetID
	}
	// Update the Geography with the allocated storage region
	if info.ResourceMetadata != nil {
		resourceMetadata.Geography = info.ResourceMetadata.Geography
	}

	creds := ""
	if environment.IsVaultEnabled() {
		creds = vault.PathForReadingKubeSecret(info.SecretRef.Namespace, info.SecretRef.Name)
	}

	request := datacatalog.CreateAssetRequest{
		ResourceMetadata:     resourceMetadata,
		Details:              details,
		Credentials:          creds,
		DestinationCatalogID: catalogID,
		DestinationAssetID:   assetID,
	}

	credentialPath := ""
	if environment.IsVaultEnabled() {
		credentialPath = vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}

	var err error
	var response *datacatalog.CreateAssetResponse
	if response, err = r.DataCatalog.CreateAsset(&request, credentialPath); err != nil {
		log.Error().Err(err).Msg("failed to receive the catalog connector response")
		return "", err
	}

	return response.AssetID, nil
}
