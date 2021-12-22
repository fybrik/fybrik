// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package v1alpha1

import (
	"fybrik.io/fybrik/pkg/model/catalog"
	// "fybrik.io/fybrik/pkg/model/catalog/"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type Asset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AssetSpec `json:"spec,omitempty"`
}

// AssetSpec defines model for AssetSpec.
type AssetSpec struct {
	// Asset details
	AssetDetails  catalog.ResourceDetails  `json:"details"`
	AssetMetadata catalog.ResourceMetadata `json:"resource"`

	// This has the vault plugin path where the data credentials will be stored as kubernetes secrets.
	// This value is assumed to be known to the catalog connector.
	// Implements::credentials
	VaultPluginPath string `json:"vaultPluginPath"`
}
