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
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["name"] = o.Name
	}

	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *Action) UnmarshalJSON(bytes []byte) (err error) {
	varConnection := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &varConnection); err == nil {
		o.Name = ActionName(varConnection["name"].(string))
		delete(varConnection, "name")
		o.AdditionalProperties = serde.Properties{Items: varConnection}
	}
	return err
}
