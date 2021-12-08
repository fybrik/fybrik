// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"io"

	datacatalogTaxonomyModels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
)

// DataCatalog is an interface of a facade to a data catalog.
type DataCatalog interface {
	GetAssetInfo(in *datacatalogTaxonomyModels.DataCatalogRequest, creds string) (*datacatalogTaxonomyModels.DataCatalogResponse, error)
	io.Closer
}
