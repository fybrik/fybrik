package clients

import (
	"io"

	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
)

// DataCatalog is an interface of a facade to a data catalog.
type DataCatalog interface {
	pb.DataCatalogServiceServer
	io.Closer
}
