// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datacatalog

import "fybrik.io/fybrik/pkg/model/taxonomy"

type GetAssetRequest struct {
	// Asset ID of the asset to be queried in the catalog
	AssetID taxonomy.AssetID `json:"assetID"`
	// Type of operation to be done on the asset
	OperationType OperationType `json:"operationType"`
}

type GetAssetResponse struct {
	// Source asset metadata like asset name, owner, geography, etc
	ResourceMetadata ResourceMetadata `json:"resourceMetadata"`
	// Source asset details like connection and data format
	Details ResourceDetails `json:"details"`
	// Vault plugin path where the data credentials will be stored as kubernetes secrets
	// This value is assumed to be known to the catalog connector.
	Credentials string `json:"credentials"`
}

type CreateAssetRequest struct {
	// The destination catalog id in which the new asset will be created based on the information provided
	// in ResourceMetadata and ResourceDetails field
	DestinationCatalogID string `json:"destinationCatalogID"`
	// +kubebuilder:validation:Optional
	// Asset ID to be used for the created asset
	DestinationAssetID string `json:"destinationAssetID,omitempty"`
	// Source asset metadata like asset name, owner, geography, etc
	ResourceMetadata ResourceMetadata `json:"resourceMetadata"`
	// Source asset details like connection and data format
	Details ResourceDetails `json:"details"`
	// +kubebuilder:validation:Optional
	// The vault plugin path where the destination data credentials will be stored as kubernetes secrets
	Credentials string `json:"credentials"`
}

type CreateAssetResponse struct {
	// The ID of the created asset based on the source asset information given in CreateAssetRequest object
	AssetID string `json:"assetID"`
}

type DeleteAssetRequest struct {
	// Asset ID of the to-be deleted asset
	AssetID taxonomy.AssetID `json:"assetID"`
}

type DeleteAssetResponse struct {
	// +kubebuilder:validation:Optional
	// The deletion status
	Status string `json:"status,omitempty"`
}

type UpdateAssetRequest struct {
	// The id of the dataset to be updated based on the information provided
	// in ResourceMetadata and ResourceDetails field
	AssetID taxonomy.AssetID `json:"assetID"`
	// Asset metadata like asset name, owner, geography, etc
	ResourceMetadata ResourceMetadata `json:"resourceMetadata"`
	// Asset details like connection and data format
	Details ResourceDetails `json:"details"`
	// +kubebuilder:validation:Optional
	// The vault plugin path where the destination data credentials will be stored as kubernetes secrets
	Credentials string `json:"credentials"`
}

type UpdateAssetResponse struct {
	// +kubebuilder:validation:Optional
	// The updation status
	Status string `json:"status,omitempty"`
}
