// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"io"
	"strings"
	"time"

	"emperror.dev/errors"

	"fybrik.io/fybrik/pkg/model/datacatalog"
)

// DataCatalog is an interface of a facade to a data catalog.
type DataCatalog interface {
	GetAssetInfo(in *datacatalog.GetAssetRequest, creds string) (*datacatalog.GetAssetResponse, error)
	CreateAsset(in *datacatalog.CreateAssetRequest, creds string) (*datacatalog.CreateAssetResponse, error)
	DeleteAsset(in *datacatalog.DeleteAssetRequest, creds string) (*datacatalog.DeleteAssetResponse, error)
	UpdateAsset(in *datacatalog.UpdateAssetRequest, creds string) (*datacatalog.UpdateAssetResponse, error)
	io.Closer
}

func NewDataCatalog(catalogProviderName, catalogConnectorAddress string, connectionTimeout time.Duration) (DataCatalog, error) {
	if strings.HasPrefix(catalogConnectorAddress, "http") {
		return NewOpenAPIDataCatalog(catalogProviderName, catalogConnectorAddress, connectionTimeout), nil
	}

	catalogClient, err := NewGrpcDataCatalog(catalogProviderName, catalogConnectorAddress, connectionTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GRPC data catalog client")
	}
	return catalogClient, nil
}
