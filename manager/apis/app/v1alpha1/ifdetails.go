// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "fybrik.io/fybrik/pkg/model/taxonomy"

// Values used in tests and for grpc connection with connectors.
// TODO(roee88): used only in tests so should be moved
const (
	S3          taxonomy.ConnectionType = "s3"
	Kafka       taxonomy.ConnectionType = "kafka"
	JdbcDb2     taxonomy.ConnectionType = "db2"
	ArrowFlight taxonomy.ConnectionType = "fybrik-arrow-flight"
	Arrow       taxonomy.DataFormat     = "arrow"
	Parquet     taxonomy.DataFormat     = "parquet"
	Table       taxonomy.DataFormat     = "table"
)

// InterfaceDetails indicate how the application or module receive or write the data
// TODO(roee88): remove redundant definition
type InterfaceDetails taxonomy.Interface
