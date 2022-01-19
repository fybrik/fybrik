// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package serde

import (
	"encoding/json"

	"github.com/mohae/deepcopy"
)

type Properties struct {
	Items map[string]interface{} `json:"-"`
}

func (in *Properties) DeepCopyInto(out *Properties) {
	// TODO: missing type assertion
	copy, _ := deepcopy.Copy(in).(*Properties)
	*out = *copy
}

func (in *Properties) DeepCopy() *Properties {
	if in == nil {
		return nil
	}
	out := new(Properties)
	in.DeepCopyInto(out)
	return out
}

func (in *Properties) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &in.Items)
	return err
}

func (in *Properties) MarshalJSON() ([]byte, error) {
	return json.Marshal(in.Items)
}
