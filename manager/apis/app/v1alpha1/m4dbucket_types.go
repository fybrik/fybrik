// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// M4DBucketSpec defines the desired state of M4DBucket
type M4DBucketSpec struct {
	// +required
	// Provisioned bucket name
	Name string `json:"name"`
	// +required
	// Endpoint
	Endpoint string `json:"endpoint"`
	// +required
	// Path where the credentials are stored
	VaultPath string `json:"vaultPath"`
}

// M4DBucketStatus defines the observed state of M4DBucket
type M4DBucketStatus struct {
	// +optional
	// Owner list: each resource is identified by namespace/name
	Owners []string `json:"owners"`
	// +optional
	// Each data asset for which the bucket is provisioned is mapped to the destination data asset (prefix)
	// This is used for sharing a single bucket by multiple applications for the same data asset
	// The data asset is identified by a string combined from dataset and catalog ids
	AssetPrefixPerDataset map[string]string `json:"assetPrefixPerDataset"`
}

// M4DBucket defines a storage asset used for implicit copy destination
// It contains endpoint, bucket name, asset name and vault path where credentials are stored
// Owner of the asset is responsible to store the credentials in vault
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type M4DBucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   M4DBucketSpec   `json:"spec,omitempty"`
	Status M4DBucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// M4DBucketList contains a list of M4DBucket
type M4DBucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []M4DBucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&M4DBucket{}, &M4DBucketList{})
}
