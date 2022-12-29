// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta2

import (
	"encoding/json"

	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
)

const typeKey = "type"

// +kubebuilder:pruning:PreserveUnknownFields
type StorageProperties struct {
	// +required
	// Storage type
	Type taxonomy.ConnectionType `json:"type"`
	// Additional storage properties
	AdditionalProperties serde.Properties `json:"-"`
}

func (o StorageProperties) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{
		typeKey: o.Type,
	}
	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *StorageProperties) UnmarshalJSON(bytes []byte) (err error) {
	items := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &items); err == nil {
		o.Type = taxonomy.ConnectionType(items[typeKey].(string))
		delete(items, typeKey)
		if len(items) == 0 {
			items = nil
		}
		o.AdditionalProperties = serde.Properties{Items: items}
	}
	return err
}
