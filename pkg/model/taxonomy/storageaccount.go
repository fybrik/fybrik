// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"fybrik.io/fybrik/pkg/serde"
)

// Reference to k8s secret holding credentials for storage access
type SecretRef struct {
	// Namespace
	Namespace string `json:"namespace"`
	// Name
	Name string `json:"name"`
}

// Properties of a shared storage account, e.g., endpoint
// +kubebuilder:pruning:PreserveUnknownFields
type StorageAccountProperties struct {
	serde.Properties `json:"-"`
}
