// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"context"
	"log"
	"net"

	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc"
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

// MockCatalogConnector returns fake data location details based on the catalog id
func MockCatalogConnector() {
	address := utils.ListeningAddress(50085)
	log.Printf("Listening on address " + address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error in listening: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDataCatalogServiceServer(s, &server{})
	pb.RegisterDataCredentialServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error in service: %v", err)
	}
}
