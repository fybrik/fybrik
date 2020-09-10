// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
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
