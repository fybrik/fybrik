// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FybrikStorageAccountSpec defines the desired state of FybrikStorageAccount
type FybrikStorageAccountSpec struct {
	// +required
	// A name of k8s secret deployed in the control plane.
	// This secret includes secretKey and accessKey credentials for S3 bucket
	SecretRef string `json:"secretRef"`
	// +required
	// Endpoint
	Endpoint string `json:"endpoint"`
	// +required
	// +kubebuilder:validation:MinItems=1
	// Regions
	Regions []string `json:"regions"`
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

	Spec   FybrikStorageAccountSpec   `json:"spec,omitempty"`
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
