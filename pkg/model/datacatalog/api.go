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
	Details          ResourceDetails  `json:"details"`
	// This has the vault plugin path where the data credentials will be stored as kubernetes secrets
	// This value is assumed to be known to the catalog connector.
	Credentials string `json:"credentials"`
}

// +fybrik:validation:object
type CreateAssetRequest struct {
	DestinationCatalogID string `json:"destinationCatalogID"`
	// +kubebuilder:validation:Optional
	DestinationAssetID string           `json:"destinationAssetID,omitempty"`
	ResourceMetadata   ResourceMetadata `json:"resourceMetadata"`
	Details            ResourceDetails  `json:"details"`
	// This has the vault plugin path where the data credentials will be stored as kubernetes secrets
	// This value is assumed to be known to the catalog connector.
	Credentials string `json:"credentials"`
}

// +fybrik:validation:object
type CreateAssetResponse struct {
	AssetID string `json:"assetID"`
}
