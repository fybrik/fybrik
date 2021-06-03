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

func (p *Arbitrary) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &p.Data); err != nil {
		return err
	}
	return nil
}

func (p *Arbitrary) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Data)
}
