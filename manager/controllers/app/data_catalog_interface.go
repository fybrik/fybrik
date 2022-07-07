// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/vault"

	"github.com/rs/zerolog/log"
)

// DeleteAsset deletes an asset from a catalog
// Input arguments:
// - assetID: DataSetID as it appears in fybrik-application
// - input: fybrik application
// Returns:
// - an error if happened
func (r *FybrikApplicationReconciler) DeleteAsset(assetID string, input *app.FybrikApplication) error {
	r.Log.Trace().Msg("DeleteAsset")
	request := datacatalog.DeleteAssetRequest{
		AssetID: taxonomy.AssetID(assetID),
	}
	credentialPath := ""
	if utils.IsVaultEnabled() {
		credentialPath = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}
	// ??? should response also be checked
	if _, err := r.DataCatalog.DeleteAsset(&request, credentialPath); err != nil {
		log.Error().Err(err).Msg("failed to receive the catalog connector response in DeleteAsset")
		return err
	}

	return nil
}

// RegisterAsset registers a new asset in the specified catalog
// Input arguments:
// - assetID: DataSetID as it appears in fybrik-application
// - catalogID: the destination catalog identifier
// - info: connection and credential details
// Returns:
// - an error if happened
// - the new asset identifier
func (r *FybrikApplicationReconciler) RegisterAsset(assetID string, catalogID string,
	info *app.DatasetDetails, input *app.FybrikApplication) (string, error) {
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
	if utils.IsVaultEnabled() {
		creds = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(info.SecretRef.Namespace, info.SecretRef.Name)
	}

	request := datacatalog.CreateAssetRequest{
		ResourceMetadata:     resourceMetadata,
		Details:              details,
		Credentials:          creds,
		DestinationCatalogID: catalogID,
		DestinationAssetID:   assetID,
	}

	credentialPath := ""
	if utils.IsVaultEnabled() {
		credentialPath = utils.GetVaultAddress() + vault.PathForReadingKubeSecret(input.Namespace, input.Spec.SecretRef)
	}

	var err error
	var response *datacatalog.CreateAssetResponse
	if response, err = r.DataCatalog.CreateAsset(&request, credentialPath); err != nil {
		log.Error().Err(err).Msg("failed to receive the catalog connector response")
		return "", err
	}

	return response.AssetID, nil
}
