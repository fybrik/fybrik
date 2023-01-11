// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta2

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
)

const typeKey = "type"
const idKey = "id"
const geographyKey = "geography"
const regionKey = "region"
const secretRefKey = "secretRef"

// FybrikStorageAccountSpec defines the desired state of FybrikStorageAccount
// +kubebuilder:pruning:PreserveUnknownFields
type FybrikStorageAccountSpec struct {
	// Identification of a storage account
	// +required
	ID string `json:"id"`
	// +required
	// A name of k8s secret deployed in the control plane.
	SecretRef string `json:"secretRef"`
	// +required
	// Type of the storage, e.g., s3
	Type taxonomy.ConnectionType `json:"type"`
	// +required
	// Storage geography
	Geography taxonomy.ProcessingLocation `json:"geography"`
	// Additional storage properties, specific to the storage type
	AdditionalProperties serde.Properties `json:"-"`
}

// FybrikStorageAccountStatus defines the observed state of FybrikStorageAccount
type FybrikStorageAccountStatus struct {
}

// FybrikStorageAccount is a storage account Fybrik uses to dynamically allocate space
// for datasets whose creation or copy it orchestrates.
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

func (o FybrikStorageAccountSpec) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{
		idKey:        o.ID,
		secretRefKey: o.SecretRef,
		typeKey:      o.Type,
		geographyKey: o.Geography,
	}
	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *FybrikStorageAccountSpec) UnmarshalJSON(bytes []byte) (err error) {
	items := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &items); err == nil {
		o.ID = items[idKey].(string)
		o.SecretRef = items[secretRefKey].(string)
		delete(items, idKey)
		delete(items, secretRefKey)
		if val, ok := items[typeKey]; ok {
			o.Type = taxonomy.ConnectionType(val.(string))
			delete(items, typeKey)
		}
		if val, ok := items[geographyKey]; ok {
			o.Geography = taxonomy.ProcessingLocation(val.(string))
			delete(items, geographyKey)
		}
		if len(items) == 0 {
			items = nil
		}
		o.AdditionalProperties = serde.Properties{Items: items}
	}
	return err
}
