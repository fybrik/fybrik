// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"context"
	"fmt"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"github.com/onsi/ginkgo"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedDataCatalogServiceServer
	pb.UnimplementedDataCredentialServiceServer
}

func (s *server) GetDatasetInfo(ctx context.Context, in *pb.CatalogDatasetRequest) (*pb.CatalogDatasetInfo, error) {
	log.Printf("Received: ")
	log.Printf("DataSetID: " + in.GetDatasetId())

	catalogID := utils.GetAttribute("catalog_id", in.GetDatasetId())
	switch catalogID {
	case "s3":
		return &pb.CatalogDatasetInfo{
			DatasetId: in.GetDatasetId(),
			Details: &pb.DatasetDetails{
				Name:       "xxx",
				DataFormat: "parquet",
				Geo:        "US",
				DataStore: &pb.DataStore{
					Type: pb.DataStore_S3,
					Name: "cos",
					S3: &pb.S3DataStore{
						Endpoint:  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
						Bucket:    "m4d-test-bucket",
						ObjectKey: "small.parq",
					},
				},
				Metadata: &pb.DatasetMetadata{},
			},
		}, nil
	case "db2":
		return &pb.CatalogDatasetInfo{
			DatasetId: in.GetDatasetId(),
			Details: &pb.DatasetDetails{
				Name:       "yyy",
				DataFormat: "table",
				Geo:        "US",
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
				Metadata: &pb.DatasetMetadata{},
			},
		}, nil
	case "kafka":
		return &pb.CatalogDatasetInfo{
			DatasetId: in.GetDatasetId(),
			Details: &pb.DatasetDetails{
				Name:       "Cars",
				DataFormat: "json",
				Geo:        "US",
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
				Metadata: &pb.DatasetMetadata{},
			},
		}, nil
	}
	return &pb.CatalogDatasetInfo{
		DatasetId: in.GetDatasetId(),
		Details: &pb.DatasetDetails{
			Name:       "yyy",
			DataFormat: "table",
			Geo:        "US",
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
			Metadata: &pb.DatasetMetadata{},
		},
	}, nil
}

func (s *server) GetCredentialsInfo(ctx context.Context, in *pb.DatasetCredentialsRequest) (*pb.DatasetCredentials, error) {
	log.Printf("Received: ")
	log.Printf("DataSetID: " + in.GetDatasetId())
	return &pb.DatasetCredentials{
		DatasetId:   in.GetDatasetId(),
		Credentials: "{\"password\":\"pswd\",\"username\":\"admin\"}",
	}, nil
}

// Creates a new mock connector or an error
func createMockCatalogConnector(port int) error {
	address := utils.ListeningAddress(port)
	log.Printf("Starting mock catalog connector on " + address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("Error when setting up mock catalog connector: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDataCatalogServiceServer(s, &server{})
	pb.RegisterDataCredentialServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("Cannot serve mock catalog connector: %v", err)
	}
	return nil
}

// MockCatalogConnector returns fake data location details based on the catalog id
func MockCatalogConnector() {
	if err := createMockCatalogConnector(50085); err != nil {
		log.Fatal(err)
	}
}

func CreateTestCatalogConnector(t ginkgo.GinkgoTInterface) {
	if err := createMockCatalogConnector(50085); err != nil {
		t.Fatal(err)
	}
}
