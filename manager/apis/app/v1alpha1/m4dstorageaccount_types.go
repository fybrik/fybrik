// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// M4DStorageAccountSpec defines the desired state of M4DStorageAccount
type M4DStorageAccountSpec struct {
	// +required
	// Secret including credentials for COS bucket
	SecretRef string `json:"secretRef"`
	// +required
	// Endpoint
	Endpoint string `json:"endpoint"`
	// +required
	// +kubebuilder:validation:MinItems=1
	// Regions
	Regions []string `json:"region"`
}

// M4DStorageAccountStatus defines the observed state of M4DStorageAccount
type M4DStorageAccountStatus struct {
}

// M4DStorageAccount defines a storage account used for copying data
// It contains endpoint, region and a reference to the credentials a
// Owner of the asset is responsible to store the credentials
// +kubebuilder:object:root=true
type M4DStorageAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   M4DStorageAccountSpec   `json:"spec,omitempty"`
	Status M4DStorageAccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// M4DStorageAccountList contains a list of M4DStorageAccount
type M4DStorageAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []M4DStorageAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&M4DStorageAccount{}, &M4DStorageAccountList{})
}
