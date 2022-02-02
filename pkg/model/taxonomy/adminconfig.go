// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"encoding/json"

	"fybrik.io/fybrik/pkg/serde"
)

type BandwidthMetric string

// +kubebuilder:pruning:PreserveUnknownFields
type StorageAccount struct {
	// Identification of a storage account
	ID string `json:"id"`
	// A name of k8s secret deployed in the control plane.
	// This secret includes secretKey and accessKey credentials for S3 bucket
	SecretRef string `json:"secretRef"`
	// Endpoint
	Endpoint string `json:"endpoint"`
	// Region
	Region ProcessingLocation `json:"region"`
	// Cost, etc.
	AdditionalProperties serde.Properties `json:"-"`
}

// +kubebuilder:pruning:PreserveUnknownFields
type Property struct {
	Name                 string           `json:"name,omitempty"`
	AdditionalProperties serde.Properties `json:"-"`
}

func (o StorageAccount) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{
		"id":        o.ID,
		"secretRef": o.SecretRef,
		"endpoint":  o.Endpoint,
		"region":    o.Region,
	}

	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}
	return json.Marshal(toSerialize)
}

func (o *StorageAccount) UnmarshalJSON(bytes []byte) (err error) {
	items := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &items); err == nil {
		o.ID = items["id"].(string)
		o.Endpoint = items["endpoint"].(string)
		o.Region = ProcessingLocation(items["region"].(string))
		o.SecretRef = items["secretRef"].(string)
		delete(items, "id")
		delete(items, "region")
		delete(items, "endpoint")
		delete(items, "secretRef")
		if len(items) == 0 {
			items = nil
		}
		o.AdditionalProperties = serde.Properties{Items: items}
	}
	return err
}

func (o Property) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{
		"name": o.Name,
	}

	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *Property) UnmarshalJSON(bytes []byte) (err error) {
	items := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &items); err == nil {
		if items["name"] != nil {
			o.Name = items["name"].(string)
			delete(items, "name")
		}
		if len(items) == 0 {
			items = nil
		}
		o.AdditionalProperties = serde.Properties{Items: items}
	}
	return err
}
