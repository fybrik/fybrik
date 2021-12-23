// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package assetmetadata

import (
	"errors"

	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
)

// DataDetails is the asset metadata and connection information received from the catalog connector
// This structure is in use by the manager and other components, such as policy manager and config policies evaluator
// TODO(roee88): this seems highly identical to the original response structure so consider dropping it
type DataDetails struct {
	// Resource metadata
	Metadata *datacatalog.ResourceMetadata `json:"metadata"`
	// Interface is the protocol and format
	Interface app.InterfaceDetails `json:"interface"`
	// Connection is the connection details in raw format as received from the connector
	Connection taxonomy.Connection `json:"connection"`
}

// Transforms a CatalogDatasetInfo into a DataDetails struct
// TODO Think about getting rid of one or the other and reuse
func CatalogDatasetToDataDetails(response *datacatalog.GetAssetResponse) (*DataDetails, error) {
	if response == nil {
		return nil, errors.New("no metadata found")
	}
	metadata := &datacatalog.ResourceMetadata{}
	response.ResourceMetadata.DeepCopyInto(metadata)
	return &DataDetails{
		Interface: app.InterfaceDetails{
			Protocol:   response.Details.Connection.Name,
			DataFormat: response.Details.DataFormat,
		},
		Connection: response.Details.Connection,
		Metadata:   metadata,
	}, nil
}
