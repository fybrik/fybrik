// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"context"
	"log"
	"net"

	"emperror.dev/errors"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
)

// This dummy catalog can serve as both a grpc server implementation that serves a dummy catalog
// with dummy datasets and dummy credentials as well as a drop in for the DataCatalog interface
// without any network traffic.
type DataCatalogDummy struct {
	pb.UnimplementedDataCatalogServiceServer
	dataDetails map[string]pb.CatalogDatasetInfo
}

func (d *DataCatalogDummy) GetDatasetInfo(ctx context.Context, req *pb.CatalogDatasetRequest) (*pb.CatalogDatasetInfo, error) {
	log.Printf("Received: ")
	log.Printf("DataSetID: " + req.GetDatasetId())

	catalogID := utils.GetAttribute("catalog_id", req.GetDatasetId())

	dataDetails, found := d.dataDetails[catalogID]
	if found {
		return &dataDetails, nil
	}

	return nil, errors.New("could not find data details")
}

func (d *DataCatalogDummy) RegisterDatasetInfo(ctx context.Context, req *pb.RegisterAssetRequest) (*pb.RegisterAssetResponse, error) {
	return nil, errors.New("functionality not yet supported")
}

func (d *DataCatalogDummy) Close() error {
	return nil
}

func NewTestCatalog() *DataCatalogDummy {
	dummyCatalog := DataCatalogDummy{
		dataDetails: make(map[string]pb.CatalogDatasetInfo),
	}
	dummyCatalog.dataDetails["s3-external"] = pb.CatalogDatasetInfo{
		DatasetId: "s3-external",
		Details: &pb.DatasetDetails{
			Name:       "xxx",
			DataFormat: "csv",
			Geo:        "neverland",
			DataStore: &pb.DataStore{
				Type: pb.DataStore_S3,
				Name: "cos",
				S3: &pb.S3DataStore{
					Endpoint:  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
					Bucket:    "m4d-test-bucket",
					ObjectKey: "test.csv",
				},
			},
			CredentialsInfo: &pb.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=m4d-system",
			},
			Metadata: &pb.DatasetMetadata{DatasetTags: []string{"PI"}},
		},
	}
	dummyCatalog.dataDetails["s3"] = pb.CatalogDatasetInfo{
		DatasetId: "s3",
		Details: &pb.DatasetDetails{
			Name:       "xxx",
			DataFormat: "parquet",
			Geo:        "theshire",
			DataStore: &pb.DataStore{
				Type: pb.DataStore_S3,
				Name: "cos",
				S3: &pb.S3DataStore{
					Endpoint:  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
					Bucket:    "m4d-test-bucket",
					ObjectKey: "small.parq",
				},
			},
			CredentialsInfo: &pb.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=m4d-system",
			},
			Metadata: &pb.DatasetMetadata{DatasetTags: []string{"PI"}},
		},
	}
	dummyCatalog.dataDetails["s3-csv"] = pb.CatalogDatasetInfo{
		DatasetId: "s3-csv",
		Details: &pb.DatasetDetails{
			Name:       "small.csv",
			DataFormat: "csv",
			Geo:        "theshire",
			DataStore: &pb.DataStore{
				Type: pb.DataStore_S3,
				Name: "cos",
				S3: &pb.S3DataStore{
					Endpoint:  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
					Bucket:    "m4d-test-bucket",
					ObjectKey: "small.csv",
				},
			},
			CredentialsInfo: &pb.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=m4d-system",
			},
			Metadata: &pb.DatasetMetadata{DatasetTags: []string{"PI"}},
		},
	}
	dummyCatalog.dataDetails["db2"] = pb.CatalogDatasetInfo{
		DatasetId: "db2",
		Details: &pb.DatasetDetails{
			Name:       "yyy",
			DataFormat: "table",
			Geo:        "theshire",
			DataStore: &pb.DataStore{
				Type: pb.DataStore_DB2,
				Name: "db2",
				Db2: &pb.Db2DataStore{
					Database: "BLUDB",
					Table:    "NQD60833.SMALL",
					Url:      "dashdb-txn-sbox-yp-lon02-02.services.eu-gb.bluemix.net",
					Port:     "50000",
					Ssl:      "false",
				},
			},
			CredentialsInfo: &pb.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=m4d-system",
			},
			Metadata: &pb.DatasetMetadata{},
		},
	}
	dummyCatalog.dataDetails["kafka"] = pb.CatalogDatasetInfo{
		DatasetId: "kafka",
		Details: &pb.DatasetDetails{
			Name:       "Cars",
			DataFormat: "json",
			Geo:        "theshire",
			DataStore: &pb.DataStore{
				Type: pb.DataStore_KAFKA,
				Name: "kafka",
				Kafka: &pb.KafkaDataStore{
					TopicName:             "topic",
					SecurityProtocol:      "SASL_SSL",
					SaslMechanism:         "SCRAM-SHA-512",
					SslTruststore:         "xyz123",
					SslTruststorePassword: "passwd",
					SchemaRegistry:        "kafka-registry",
					BootstrapServers:      "http://kafka-servers",
					KeyDeserializer:       "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer",
					ValueDeserializer:     "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer",
				},
			},
			CredentialsInfo: &pb.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=m4d-system",
			},
			Metadata: &pb.DatasetMetadata{},
		},
	}

	return &dummyCatalog
}

var connector *grpc.Server = nil

// Creates a new mock connector or an error
func createMockCatalogConnector(port int) error {
	if connector != nil {
		return errors.New("a catalog connector was already started")
	}
	address := utils.ListeningAddress(port)
	log.Printf("Starting mock catalog connector on " + address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrap(err, "failed setting up mock catalog connector")
	}
	s := grpc.NewServer()
	connector = s
	dummyCatalog := NewTestCatalog()
	pb.RegisterDataCatalogServiceServer(s, dummyCatalog)
	if err := s.Serve(lis); err != nil {
		return errors.Wrap(err, "failed in serve of mock catalog connector")
	}
	return nil
}

// MockCatalogConnector returns fake data location details based on the catalog id
func MockCatalogConnector() {
	if err := createMockCatalogConnector(8080); err != nil {
		log.Fatal(err)
	}
}
