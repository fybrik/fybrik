package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"

	"emperror.dev/errors"
	"google.golang.org/grpc"
)

// Ensure that grpcDataCatalog implements the DataCatalog interface
var _ DataCatalog = (*grpcDataCatalog)(nil)

type grpcDataCatalog struct {
	name       string
	connection *grpc.ClientConn
	client     pb.DataCatalogServiceClient
}

// NewGrpcDataCatalog creates a DataCatalog facade that connects to a GRPC service
// You must call .Close() when you are done using the created instance
func NewGrpcDataCatalog(name string, connectionURL string, connectionTimeout time.Duration) (DataCatalog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	connection, err := grpc.DialContext(ctx, connectionURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("NewGrpcDataCatalog failed when connecting to %s", connectionURL))
	}
	return &grpcDataCatalog{
		name:       name,
		client:     pb.NewDataCatalogServiceClient(connection),
		connection: connection,
	}, nil
}

func (m *grpcDataCatalog) GetDatasetInfo(ctx context.Context, in *pb.CatalogDatasetRequest) (*pb.CatalogDatasetInfo, error) {
	result, err := m.client.GetDatasetInfo(ctx, in)
	return result, errors.Wrap(err, fmt.Sprintf("get dataset info from %s failed", m.name))
}

func (m *grpcDataCatalog) RegisterDatasetInfo(ctx context.Context, in *pb.RegisterAssetRequest) (*pb.RegisterAssetResponse, error) {
	result, err := m.client.RegisterDatasetInfo(ctx, in)
	return result, errors.Wrap(err, fmt.Sprintf("register dataset info in %s failed", m.name))
}

func (m *grpcDataCatalog) Close() error {
	return m.connection.Close()
}
