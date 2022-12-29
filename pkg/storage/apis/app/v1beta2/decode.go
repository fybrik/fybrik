// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta2

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func (o *FybrikStorageAccount) DecodeYaml(bytes []byte) error {
	if err := yaml.Unmarshal(bytes, o); err != nil {
		return err
	}
	// get additional properties
	object := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(bytes, &object); err != nil {
		return err
	}
	spec := object.UnstructuredContent()["spec"]
	specData, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return o.Spec.Properties.UnmarshalJSON(specData)
}
