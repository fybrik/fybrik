// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datacatalog

import "fybrik.io/fybrik/pkg/model/taxonomy"

// Type of operation requested for the asset
type OperationType string

// List of operationTypes
const (
	READ OperationType = "read"
)

// ResourceMetadata defines model for resource metadata
type ResourceMetadata struct {
	// Name of the resource
	Name string `json:"name,omitempty"`
	// Owner of the resource
	Owner string `json:"owner,omitempty"`
	// Geography of the resource
	Geography string `json:"geography,omitempty"`
	// Tags associated with the asset
	Tags *taxonomy.Tags `json:"tags,omitempty"`
	// Columns associated with the asset
	Columns []ResourceColumn `json:"columns,omitempty"`
}

// ResourceColumn represents a column in a tabular resource
type ResourceColumn struct {
	// Name of the column
	Name string `json:"name"`
	// Tags associated with the column
	Tags *taxonomy.Tags `json:"tags,omitempty"`
}

// ResourceDetails includes asset connection details
type ResourceDetails struct {
	// Connection information
	Connection taxonomy.Connection `json:"connection"`
	// Data format
	DataFormat taxonomy.DataFormat `json:"dataFormat,omitempty"`
}
