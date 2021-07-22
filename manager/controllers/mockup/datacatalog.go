// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
)

// This dummy catalog can serve as both a grpc server implementation that serves a dummy catalog
// with dummy datasets and dummy credentials as well as a drop in for the DataCatalog interface
// without any network traffic.
type DataCatalogDummy struct {
	pb.UnimplementedDataCatalogServiceServer
	dataDetails map[string]pb.CatalogDatasetInfo
}

func (d *DataCatalogDummy) GetDatasetInfo(ctx context.Context, in *pb.CatalogDatasetRequest) (*pb.CatalogDatasetInfo, error) {
	datasetID := in.GetDatasetId()
	log.Printf("MockDataCatalog.GetDatasetInfo called with DataSetID " + datasetID)

	splittedID := strings.SplitN(datasetID, "/", 2)
	if len(splittedID) != 2 {
		panic(fmt.Sprintf("Invalid dataset ID for mock: %s", datasetID))
	}

	catalogID := splittedID[0]

	dataDetails, found := d.dataDetails[catalogID]
	if found {
		return &dataDetails, nil
	}

	return nil, errors.New("could not find data details")
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
					Bucket:    "fybrik-test-bucket",
					ObjectKey: "test.csv",
				},
			},
			CredentialsInfo: &pb.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
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
					Bucket:    "fybrik-test-bucket",
					ObjectKey: "small.parq",
				},
			},
			CredentialsInfo: &pb.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
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
					Bucket:    "fybrik-test-bucket",
					ObjectKey: "small.csv",
				},
			},
			CredentialsInfo: &pb.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
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
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
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
				VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
			},
			Metadata: &pb.DatasetMetadata{},
		},
	}

	return &dummyCatalog
}
