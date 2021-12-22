package taxonomy

import (
	"encoding/json"

	"fybrik.io/fybrik/pkg/serde"
)

type AssetID string

type ConnectionType string

// +kubebuilder:pruning:PreserveUnknownFields
type Connection struct {
	Name                 ConnectionType   `json:"name"`
	AdditionalProperties serde.Properties `json:"-"`
}

type DataFormat string

type Interface struct {
	// Protocol defines the interface protocol used for data transactions
	Protocol ConnectionType `json:"protocol"` // TODO(roee88): should this be named ConnectionType instead of Protocol
	// DataFormat defines the data format type
	DataFormat DataFormat `json:"dataformat,omitempty"`
}

// +kubebuilder:pruning:PreserveUnknownFields
type Tags struct {
	serde.Properties `json:"-"`
}

func (o Connection) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["name"] = o.Name
	}

	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *Connection) UnmarshalJSON(bytes []byte) (err error) {
	varConnection := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &varConnection); err == nil {
		o.Name = ConnectionType(varConnection["name"].(string))
		delete(varConnection, "name")
		o.AdditionalProperties = serde.Properties{Items: varConnection}
	}
	return err
}
