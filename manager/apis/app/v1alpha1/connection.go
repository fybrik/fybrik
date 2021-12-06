// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
)

// +kubebuilder:validation:Type=object
// +kubebuilder:pruning:PreserveUnknownFields
type ConnectionDetails struct {
	taxonomymodels.Connection
}

func (connection *ConnectionDetails) UnmarshalJSON(data []byte) error {
	return connection.Connection.UnmarshalJSON(data)
}

func (connection *ConnectionDetails) MarshalJSON() ([]byte, error) {
	return connection.Connection.MarshalJSON()
}
