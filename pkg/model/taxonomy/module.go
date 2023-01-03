// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"encoding/json"

	"fybrik.io/fybrik/pkg/serde"
)

// Capability declared by the module, e.g., read, delete, copy, write, transform
type Capability string

// Type of the plugin, not supported yet
type PluginType string

// Name of the action to be performed, or Deny if access to the data is forbidden
// Action names should be defined in additional taxonomy layers
type ActionName string

// DataFlow indicates how the data is used by the workload, e.g., it is being read, copied, written or deleted
// +kubebuilder:validation:Enum=read;write;delete;copy
type DataFlow string

const (
	// ReadFlow indicates a data set is being read
	ReadFlow DataFlow = "read"

	// WriteFlow indicates a data set is being written
	WriteFlow DataFlow = "write"

	// DeleteFlow indicates a data set is being deleted
	DeleteFlow DataFlow = "delete"

	// CopyFlow indicates a data set is being copied
	CopyFlow DataFlow = "copy"
)

// Action to be performed on the data, e.g., masking
// +kubebuilder:pruning:PreserveUnknownFields
type Action struct {
	// Action name
	Name ActionName `json:"name"`
	// Action properties, e.g., names of columns that should be masked
	AdditionalProperties serde.Properties `json:"-"`
}

func (o Action) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{
		nameKey: o.Name,
	}

	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *Action) UnmarshalJSON(bytes []byte) (err error) {
	items := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &items); err == nil {
		o.Name = ActionName(items[nameKey].(string))
		delete(items, nameKey)
		if len(items) == 0 {
			items = nil
		}
		o.AdditionalProperties = serde.Properties{Items: items}
	}
	return err
}
