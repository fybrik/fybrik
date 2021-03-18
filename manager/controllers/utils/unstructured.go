// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/json"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CreateUnstructured creates a new Unstructured runtime object
func CreateUnstructured(group, version, kind, name, namespace string) *unstructured.Unstructured {
	result := &unstructured.Unstructured{}
	result.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	})
	result.SetName(name)
	result.SetNamespace(namespace)
	return result
}

// UnstructuredAsLabels is an implementation of labels.Labels interface
// which allows us to take advantage of k8s labels library
// for the purposes of evaluating fail and success conditions
type UnstructuredAsLabels struct {
	Data *unstructured.Unstructured
}

// Has returns whether the provided label exists.
func (c UnstructuredAsLabels) Has(label string) bool {
	obj := c.Data.UnstructuredContent()
	fields := strings.Split(label, ".")
	// value is not returned
	_, exists, err := unstructured.NestedFieldNoCopy(obj, fields...)
	if err != nil || !exists {
		return false
	}
	return true
}

// Get returns the value for the provided label.
func (c UnstructuredAsLabels) Get(label string) string {
	obj := c.Data.UnstructuredContent()
	fields := strings.Split(label, ".")
	// not checking whether the label exists and is valid. Assuming Get is called after Has in labels package evaluation
	// fetch a string value
	if valStr, _, err := unstructured.NestedString(obj, fields...); err == nil {
		return valStr
	}
	// convert a received interface into a string
	val, _, err := unstructured.NestedFieldNoCopy(obj, fields...)
	if err != nil {
		return ""
	}
	valToBytes, err := json.Marshal(val)
	if err != nil {
		return ""
	}
	return string(valToBytes)
}
