// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"encoding/json"

	"fybrik.io/fybrik/pkg/serde"
)

type Capability string

type PluginType string

type ActionName string

// +kubebuilder:pruning:PreserveUnknownFields
type Action struct {
	Name                 ActionName       `json:"name"`
	AdditionalProperties serde.Properties `json:"-"`
}

func (o Action) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{
		"name": o.Name,
	}

	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *Action) UnmarshalJSON(bytes []byte) (err error) {
	items := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &items); err == nil {
		o.Name = ActionName(items["name"].(string))
		delete(items, "name")
		if len(items) == 0 {
			items = nil
		}
		o.AdditionalProperties = serde.Properties{Items: items}
	}
	return err
}
