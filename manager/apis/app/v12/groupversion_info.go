// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

// Package v12 contains API Schema definitions for the api v12 API group
// +kubebuilder:object:generate=true
// +groupName=app.fybrik.io
package v12

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "app.fybrik.io", Version: "v12"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
