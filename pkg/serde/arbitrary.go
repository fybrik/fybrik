package serde

import (
	"encoding/json"

	"github.com/mohae/deepcopy"
)

// +kubebuilder:validation:Type=object
// +kubebuilder:pruning:PreserveUnknownFields
type Arbitrary struct {
	Data interface{} `json:"-"`
}

func NewArbitrary(in interface{}) *Arbitrary {
	return &Arbitrary{
		Data: in,
	}
}

func (in *Arbitrary) DeepCopyInto(out *Arbitrary) {
	// TODO: missing type assertion
	copy, _ := deepcopy.Copy(in).(*Arbitrary)
	*out = *copy
}

func (in *Arbitrary) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &in.Data); err != nil {
		return err
	}
	return nil
}

func (in *Arbitrary) MarshalJSON() ([]byte, error) {
	return json.Marshal(in.Data)
}

func (in *Arbitrary) Into(target interface{}) error {
	raw, err := json.Marshal(in.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
