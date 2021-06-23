// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// Values used in tests and for grpc connection with connectors.
const (
	S3          string = "s3"
	Kafka       string = "kafka"
	JdbcDb2     string = "jdbc-db2"
	ArrowFlight string = "m4d-arrow-flight"
	Arrow       string = "arrow"
	Parquet     string = "parquet"
	Table       string = "table"
)

// InterfaceDetails indicate how the application or module receive or write the data
type InterfaceDetails struct {
	// Protocol defines the interface protocol used for data transactions
	// +required
	Protocol string `json:"protocol"`

	// DataFormat defines the data format type
	// +optional
	DataFormat string `json:"dataformat,omitempty"` // To be removed in future
}
