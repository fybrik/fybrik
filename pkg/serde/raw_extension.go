package serde

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
)

func ToRawExtension(value interface{}) (*runtime.RawExtension, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{Raw: raw}, nil
}

func FromRawExtention(ext runtime.RawExtension, out interface{}) error {
	return json.Unmarshal(ext.Raw, out)
}
