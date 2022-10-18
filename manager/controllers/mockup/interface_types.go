// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// connection types and formats used in tests and mockup connectors
const (
	S3          taxonomy.ConnectionType = "s3"
	Kafka       taxonomy.ConnectionType = "kafka"
	JdbcDB2     taxonomy.ConnectionType = "db2"
	ArrowFlight taxonomy.ConnectionType = "fybrik-arrow-flight"

	Parquet taxonomy.DataFormat = "parquet"
	CSV     taxonomy.DataFormat = "csv"
)
