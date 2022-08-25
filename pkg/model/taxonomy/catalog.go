// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"encoding/json"

	"fybrik.io/fybrik/pkg/serde"
)

const nameKey = "name"

// Asset ID of the asset to be queried in the catalog
type AssetID string

// location information
type ProcessingLocation string

// Name of the connection type to the data source
type ConnectionType string

// +kubebuilder:pruning:PreserveUnknownFields
// Name of the connection to the data source
type Connection struct {
	// Name of the connection to the data source
	Name                 ConnectionType   `json:"name"`
	AdditionalProperties serde.Properties `json:"-"`
}

// Format of the data
type DataFormat string

type Interface struct {
	// Protocol defines the interface protocol used for data transactions
	Protocol ConnectionType `json:"protocol"` // TODO(roee88): should this be named ConnectionType instead of Protocol
	// DataFormat defines the data format type
	DataFormat DataFormat `json:"dataformat,omitempty"`
}

// +kubebuilder:pruning:PreserveUnknownFields
// Additional metadata for the asset/field
type Tags struct {
	serde.Properties `json:"-"`
}

func (o Connection) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{
		nameKey: o.Name,
	}

	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *Connection) UnmarshalJSON(bytes []byte) (err error) {
	items := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &items); err == nil {
		o.Name = ConnectionType(items[nameKey].(string))
		delete(items, nameKey)
		if len(items) == 0 {
			items = nil
		}
		o.AdditionalProperties = serde.Properties{Items: items}
	}
	return err
}
