// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// IFProtocol defines interface protocol for data transactions
type IFProtocol string

// DataFormatType defines data format type
type DataFormatType string

// DataFormatType valid values
const (
	Parquet DataFormatType = "parquet"
	Table   DataFormatType = "table" // remove?
	CSV     DataFormatType = "csv"
	JSON    DataFormatType = "json"
	AVRO    DataFormatType = "avro"
	ORC     DataFormatType = "orc"
	Binary  DataFormatType = "binary"
	Arrow   DataFormatType = "arrow"
)

// IFProtocol valid values
const (
	S3          IFProtocol = "s3"
	Kafka       IFProtocol = "kafka"
	JdbcDb2     IFProtocol = "jdbc-db2"
	ArrowFlight IFProtocol = "m4d-arrow-flight"
)

// InterfaceDetails indicate how the application or module receive or write the data
type InterfaceDetails struct {

	// +required
	Protocol IFProtocol `json:"protocol"`

	// +optional
	DataFormat DataFormatType `json:"dataformat,omitempty"` // To be removed in future
}
