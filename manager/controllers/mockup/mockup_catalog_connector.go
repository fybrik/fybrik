// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"

	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"github.com/onsi/ginkgo"
	"google.golang.org/grpc"
)

type DataCatalogDummy struct {
	credentials map[string]pb.Credentials
	dataDetails map[string]pb.CatalogDatasetInfo
}

func (d *DataCatalogDummy) GetConnectionDetails(req *modules.DataInfo, input *app.M4DApplication) error {
	return d.GetConnectionDetailsByID(req, req.Context.DataSetID)
}

func (d *DataCatalogDummy) GetCredentials(req *modules.DataInfo, input *app.M4DApplication) error {
	return d.GetCredentialsByID(req, req.Context.DataSetID)
}

func (d *DataCatalogDummy) GetConnectionDetailsByID(req *modules.DataInfo, datasetID string) error {
	dataDetails, found := d.dataDetails[datasetID]
	if found {
		details, err := modules.CatalogDatasetToDataDetails(&dataDetails)
		if err != nil {
			return err
		}
		req.DataDetails = details
		return nil
	}
	return errors.New("could not find data details")
}

func (d *DataCatalogDummy) GetCredentialsByID(req *modules.DataInfo, datasetID string) error {
	credentials, found := d.credentials[datasetID]
	if found {
		req.Credentials = &pb.DatasetCredentials{
			DatasetId: req.Context.DataSetID,
			Creds:     &credentials,
		}
		return nil
	}
	return errors.New("could not find credentials")
}

func NewTestCatalog() *DataCatalogDummy {
	dummyCatalog := DataCatalogDummy{
		credentials: make(map[string]pb.Credentials),
		dataDetails: make(map[string]pb.CatalogDatasetInfo),
	}
	dummyCatalog.dataDetails["s3-external"] = pb.CatalogDatasetInfo{
		DatasetId: "s3-external",
		Details: &pb.DatasetDetails{
			Name:       "xxx",
			DataFormat: "parquet",
			Geo:        "Germany",
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
	dummyCatalog.credentials["s3-csv"] = pb.Credentials{Username: "admin", Password: "pswd"}

	return &dummyCatalog
}

type server struct {
	pb.UnimplementedDataCatalogServiceServer
	pb.UnimplementedDataCredentialServiceServer
	c *DataCatalogDummy
}

func (s *server) GetDatasetInfo(ctx context.Context, in *pb.CatalogDatasetRequest) (*pb.CatalogDatasetInfo, error) {
	log.Printf("Received: ")
	log.Printf("DataSetID: " + in.GetDatasetId())

	catalogID := utils.GetAttribute("catalog_id", in.GetDatasetId())

	dataDetails, found := s.c.dataDetails[catalogID]
	if found {
		return &dataDetails, nil
	}
	return nil, errors.New("catalog item not found")
}

//TODO: remove this!
func (s *server) GetCredentialsInfo(ctx context.Context, in *pb.DatasetCredentialsRequest) (*pb.DatasetCredentials, error) {
	log.Printf("Received: ")
	log.Printf("DataSetID: " + in.GetDatasetId())
	return &pb.DatasetCredentials{
		DatasetId: in.GetDatasetId(),
		Creds:     &pb.Credentials{Username: "admin", Password: "pswd"},
	}, nil
}

//TODO: remove this!
func (s *server) RegisterDatasetInfo(ctx context.Context, in *pb.RegisterAssetRequest) (*pb.RegisterAssetResponse, error) {
	return &pb.RegisterAssetResponse{AssetId: "NewAsset"}, nil
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
		return fmt.Errorf("Error when setting up mock catalog connector: %v", err)
	}
	s := grpc.NewServer()
	connector = s
	dummyCatalog := NewTestCatalog()
	pb.RegisterDataCatalogServiceServer(s, &server{c: dummyCatalog})
	pb.RegisterDataCredentialServiceServer(s, &server{c: dummyCatalog})
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("Cannot serve mock catalog connector: %v", err)
	}
	return nil
}

// MockCatalogConnector returns fake data location details based on the catalog id
func MockCatalogConnector() {
	if err := createMockCatalogConnector(8080); err != nil {
		log.Fatal(err)
	}
}

func CreateTestCatalogConnector(t ginkgo.GinkgoTInterface) {
	if err := createMockCatalogConnector(50085); err != nil {
		t.Fatal(err)
	}
}

func KillServer() {
	if connector != nil {
		log.Print("Killing server...")
		connector.Stop()
		connector = nil
	}
}
