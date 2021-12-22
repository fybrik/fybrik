package taxonomy

import (
	"encoding/json"

	"fybrik.io/fybrik/pkg/serde"
)

// +kubebuilder:pruning:PreserveUnknownFields
type PolicyManagerRequestContext struct {
	serde.Properties `json:"-"`
}

type PolicyManagerActionName string

// +kubebuilder:pruning:PreserveUnknownFields
type PolicyManagerAction struct {
	Name                 PolicyManagerActionName `json:"name"`
	AdditionalProperties serde.Properties        `json:"-"`
}

func (o PolicyManagerAction) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["name"] = o.Name
	}

	for key, value := range o.AdditionalProperties.Items {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *PolicyManagerAction) UnmarshalJSON(bytes []byte) (err error) {
	varConnection := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &varConnection); err == nil {
		o.Name = PolicyManagerActionName(varConnection["name"].(string))
		delete(varConnection, "name")
		o.AdditionalProperties = serde.Properties{Items: varConnection}
	}
	return err
}
