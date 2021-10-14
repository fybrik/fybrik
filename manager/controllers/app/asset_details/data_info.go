// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package asset_details

import (
	"fybrik.io/fybrik/pkg/serde"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
)

// DataDetails is the information received from the catalog connector
type DataDetails struct {
	// Name of the asset
	Name string
	// Interface is the protocol and format
	Interface app.InterfaceDetails
	// Geography is the geo-location of the asset
	Geography string
	// Connection is the connection details in raw format as received from the connector
	Connection serde.Arbitrary
	// Metadata
	Metadata *pb.DatasetMetadata
}
