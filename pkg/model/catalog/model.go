package catalog

import "fybrik.io/fybrik/pkg/model/taxonomy"

// OperationType Type of operation requested for the asset
// +kubebuilder:validation:Enum=read;
type OperationType string

// List of operationType
const (
	READ OperationType = "read"
)

// ResourceMetadata defines model for resource metadata
type ResourceMetadata struct {
	// Name of the resource
	Name string `json:"name"`
	// Owner of the resource
	// +optional
	Owner *string `json:"owner,omitempty"`
	// Geography of the resource
	// +optional
	Geography *string `json:"geography,omitempty"`
	// Tags associated with the asset
	// +optional
	Tags taxonomy.Tags `json:"tags,omitempty"`
	// Columns associated with the asset
	// +optional
	Columns *[]ResourceColumn `json:"columns,omitempty"`
}

// ResourceColumn represents a column in a tabular resource
type ResourceColumn struct {
	// Name of the column
	Name string `json:"name"`
	// Tags associated with the column
	// +optional
	Tags taxonomy.Tags `json:"tags,omitempty"`
}

// ResourceDetails includes asset connection details
type ResourceDetails struct {
	// Connection information
	Connection taxonomy.Connection `json:"connection"`
	// Data format
	// +optional
	DataFormat taxonomy.DataFormat `json:"dataFormat,omitempty"`
}
