// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// FybrikStorageAccountSpec defines the desired state of FybrikStorageAccount
type FybrikStorageAccountSpec struct {
	// Identification of a storage account
	// +required
	ID string `json:"id"`
	// Connection type
	// +required
	Type taxonomy.ConnectionType `json:"type"`
	// +optional
	// A name of k8s secret deployed in the control plane.
	// This secret includes credentials required for the connection
	SecretRef string `json:"secretRef,omitempty"`
	// +required
	// Storage geography
	Geography taxonomy.ProcessingLocation `json:"geography"`
	// +required
	// Connection properties
	Properties map[string]string `json:"properties"`
}

// FybrikStorageAccountStatus defines the observed state of FybrikStorageAccount
type FybrikStorageAccountStatus struct {
}

// FybrikStorageAccount defines a storage account used for copying data.
// It contains connection details of the shared storage and refers to the secret that stores storage credentials.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
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
