// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package assetmetadata

import (
	"errors"

	"fybrik.io/fybrik/pkg/serde"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	dc "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
)

// DataDetails is the asset metadata and connection information received from the catalog connector
// This structure is in use by the manager and other components, such as policy manager and config policies evaluator
type DataDetails struct {
	// Resource metadata
	Metadata *dc.Resource `json:"metadata"`
	// Interface is the protocol and format
	Interface app.InterfaceDetails `json:"interface"`
	// Connection is the connection details in raw format as received from the connector
	Connection serde.Arbitrary `json:"connection"`
}

// Transforms a CatalogDatasetInfo into a DataDetails struct
// TODO Think about getting rid of one or the other and reuse
func CatalogDatasetToDataDetails(response *dc.DataCatalogResponse) (*DataDetails, error) {
	if response == nil {
		return nil, errors.New("no metadata found")
	}
	connection := serde.NewArbitrary(response.Details.Connection)
	metadata := &dc.Resource{}
	response.ResourceMetadata.DeepCopyInto(metadata)
	return &DataDetails{
		Interface: app.InterfaceDetails{
			Protocol:   response.Details.Connection.Name,
			DataFormat: *response.Details.DataFormat,
		},
		Connection: *connection,
		Metadata:   metadata,
	}, nil
}
