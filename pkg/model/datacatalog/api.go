// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datacatalog

import "fybrik.io/fybrik/pkg/model/taxonomy"

// +fybrik:validation:object
type GetAssetRequest struct {
	AssetID       taxonomy.AssetID `json:"assetID"`
	OperationType OperationType    `json:"operationType"`
}

// +fybrik:validation:object
type GetAssetResponse struct {
	ResourceMetadata ResourceMetadata `json:"resourceMetadata"`
	ResourceDetails  ResourceDetails  `json:"resourceDetails"`
	// This has the vault plugin path where the data credentials will be stored as kubernetes secrets
	// This value is assumed to be known to the catalog connector.
	Credentials string `json:"credentials"`
}

// +fybrik:validation:object
type CreateAssetRequest struct {
	// This has the information about the destination catalog id that new asset that will be created with the information provided in ResourceMetadata and Details field will be stored.
	DestinationCatalogID string `json:"destinationCatalogID"`
	// +kubebuilder:validation:Optional
	// This is an optional field provided to give information about the asset id to be used for the created asset.
	DestinationAssetID string `json:"destinationAssetID,omitempty"`
	// This field has the information about the source asset metadata like asset name, owner, geography, etc
	ResourceMetadata ResourceMetadata `json:"resourceMetadata"`
	// This field has more details about the source asset like connection and dataformat
	ResourceDetails ResourceDetails `json:"resourceDetails"`
	// +kubebuilder:validation:Optional
	// This optional field has the vault plugin path where the destination data credentials will be stored as kubernetes secrets
	Credentials string `json:"credentials"`
}

// +fybrik:validation:object
type CreateAssetResponse struct {
	// This field stores the newly created asset id based on the source asset information given in CreateAssetRequest object.
	AssetID string `json:"assetID"`
}
