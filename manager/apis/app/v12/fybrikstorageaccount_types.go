// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v12

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// FybrikStorageAccountSpec defines the desired state of FybrikStorageAccount
type FybrikStorageAccountSpec struct {
	// Identification of a storage account
	// +required
	ID string `json:"id"`
	// +required
	// A name of k8s secret deployed in the control plane.
	// This secret includes secretKey and accessKey credentials for S3 bucket
	SecretRef string `json:"secretRef"`
	// +required
	// Storage region
	Region taxonomy.ProcessingLocation `json:"region"`
	// +required
	// Endpoint for accessing the data
	Endpoint string `json:"endpoint"`
}

// FybrikStorageAccountStatus defines the observed state of FybrikStorageAccount
type FybrikStorageAccountStatus struct {
}

// FybrikStorageAccount defines a storage account used for copying data.
// Only S3 based storage is supported.
// It contains endpoint, region and a reference to the credentials a
// Owner of the asset is responsible to store the credentials
// +kubebuilder:object:root=true
type FybrikStorageAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   FybrikStorageAccountSpec   `json:"spec"`
	Status FybrikStorageAccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FybrikStorageAccountList contains a list of FybrikStorageAccount
type FybrikStorageAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FybrikStorageAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FybrikStorageAccount{}, &FybrikStorageAccountList{})
}
