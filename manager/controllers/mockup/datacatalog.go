// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	catalogmodels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
)

type DataCatalogDummy struct {
	dataDetails map[string]catalogmodels.DataCatalogResponse
}

func (d *DataCatalogDummy) GetAssetInfo(in *catalogmodels.DataCatalogRequest, creds string) (*catalogmodels.DataCatalogResponse, error) {
	datasetID := in.AssetID
	log.Printf("MockDataCatalog.GetDatasetInfo called with DataSetID " + datasetID)

	splittedID := strings.SplitN(datasetID, "/", 2)
	if len(splittedID) != 2 {
		panic(fmt.Sprintf("Invalid dataset ID for mock: %s", datasetID))
	}

	catalogID := splittedID[0]

	dataDetails, found := d.dataDetails[catalogID]
	if found {
		log.Printf("GetAssetInfo in DataCatalogDummy returns:")
		responseBytes, errJSON := json.MarshalIndent(dataDetails, "", "\t")
		if errJSON != nil {
			return nil, fmt.Errorf("error in GetAssetInfo in DataCatalogDummy: %v", errJSON)
		}
		log.Print(string(responseBytes))
		return &dataDetails, nil
	}

	return nil, errors.New("could not find data details")
}

func (d *DataCatalogDummy) Close() error {
	return nil
}

func NewTestCatalog() *DataCatalogDummy {
	dummyCatalog := DataCatalogDummy{
		dataDetails: make(map[string]catalogmodels.DataCatalogResponse),
	}

	tags := make(map[string]interface{})
	tags["tags"] = []string{"PI"}

	var connection catalogmodels.Connection
	var connectionStr = `{
		"name": "s3",
		"s3": {
			"Endpoint":  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
			"Bucket":    "fybrik-test-bucket",
			"ObjectKey": "test.csv"
		}}`
	err := json.Unmarshal([]byte(connectionStr), &connection)
	if err != nil {
		panic(err)
	}
	// connectionBytes, errJSON := json.MarshalIndent(connection, "", "\t")
	// if errJSON != nil {
	// 	return nil, fmt.Errorf("error Marshalling External Catalog Connector Response: %v", errJSON)
	// }
	// err := json.Unmarshal(actionBytes, &actionOnCols)
	// if err != nil {
	// 	return nil, fmt.Errorf("error in unmarshalling actionBytes : %v", err)
	// }
	geo := "neverland"
	dataformat := "csv"
	dummyCatalog.dataDetails["s3-external"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "s3-external",
			Geography: &geo,
			Tags:      &tags,
		},
		Details: catalogmodels.Details{
			Connection: connection,
			DataFormat: &dataformat},
		Credentials: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	}

	// pb.CatalogDatasetInfo{
	// 	DatasetId: "s3-external",
	// 	Details: &pb.DatasetDetails{
	// 		Name:       "xxx",
	// 		DataFormat: "csv",
	// 		Geo:        "neverland",
	// 		DataStore: &pb.DataStore{
	// 			Type: pb.DataStore_S3,
	// 			Name: "cos",
	// 			S3: &pb.S3DataStore{
	// 				Endpoint:  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
	// 				Bucket:    "fybrik-test-bucket",
	// 				ObjectKey: "test.csv",
	// 			},
	// 		},
	// 		CredentialsInfo: &pb.CredentialsInfo{
	// 			VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	// 		},
	// 		Metadata: &pb.DatasetMetadata{DatasetTags: []string{"PI"}},
	// 	},
	// }
	geo2 := "theshire"
	dataformat2 := "parquet"
	var connection2 catalogmodels.Connection
	var connectionStr2 = `{
		"name": "s3",
		"s3": {
			"Endpoint":  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
	 		"Bucket":    "fybrik-test-bucket",
	 		"ObjectKey": "small.parq"
			}}`
	err = json.Unmarshal([]byte(connectionStr2), &connection2)
	if err != nil {
		panic(err)
	}

	dummyCatalog.dataDetails["s3"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "s3",
			Geography: &geo2,
			Tags:      &tags,
		},
		Details: catalogmodels.Details{
			Connection: connection2,
			DataFormat: &dataformat2},
		Credentials: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	}

	// pb.CatalogDatasetInfo{
	// 	DatasetId: "s3",
	// 	Details: &pb.DatasetDetails{
	// 		Name:       "xxx",
	// 		DataFormat: "parquet",
	// 		Geo:        "theshire",
	// 		DataStore: &pb.DataStore{
	// 			Type: pb.DataStore_S3,
	// 			Name: "cos",
	// 			S3: &pb.S3DataStore{
	// 				Endpoint:  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
	// 				Bucket:    "fybrik-test-bucket",
	// 				ObjectKey: "small.parq",
	// 			},
	// 		},
	// 		CredentialsInfo: &pb.CredentialsInfo{
	// 			VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	// 		},
	// 		Metadata: &pb.DatasetMetadata{DatasetTags: []string{"PI"}},
	// 	},
	// }

	geo3 := "theshire"
	dataformat3 := "csv"
	var connection3 catalogmodels.Connection
	var connectionStr3 = `{
		"name": "s3",
		"s3": {
			"Endpoint":  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
			"Bucket":    "fybrik-test-bucket",
			"ObjectKey": "small.csv"
			}
		}`
	err = json.Unmarshal([]byte(connectionStr3), &connection3)
	if err != nil {
		panic(err)
	}

	dummyCatalog.dataDetails["s3-csv"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "small.csv",
			Geography: &geo3,
			Tags:      &tags,
		},
		Details: catalogmodels.Details{
			Connection: connection3,
			DataFormat: &dataformat3},
		Credentials: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	}

	// dummyCatalog.dataDetails["s3-csv"] = pb.CatalogDatasetInfo{
	// 	DatasetId: "s3-csv",
	// 	Details: &pb.DatasetDetails{
	// 		Name:       "small.csv",
	// 		DataFormat: "csv",
	// 		Geo:        "theshire",
	// 		DataStore: &pb.DataStore{
	// 			Type: pb.DataStore_S3,
	// 			Name: "cos",
	// 			S3: &pb.S3DataStore{
	// 				Endpoint:  "s3.eu-gb.cloud-object-storage.appdomain.cloud",
	// 				Bucket:    "fybrik-test-bucket",
	// 				ObjectKey: "small.csv",
	// 			},
	// 		},
	// 		CredentialsInfo: &pb.CredentialsInfo{
	// 			VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	// 		},
	// 		Metadata: &pb.DatasetMetadata{DatasetTags: []string{"PI"}},
	// 	},
	// }

	geo4 := "theshire"
	dataformat4 := "table"
	var connection4 catalogmodels.Connection
	var connectionStr4 = `{
		"name": "DB2",
		"db2": {
			"Database": "BLUDB",
			"Table":    "NQD60833.SMALL",
			"Url":      "dashdb-txn-sbox-yp-lon02-02.services.eu-gb.bluemix.net",
			"Port":     "50000",
			"Ssl":      "false"
			}
		}`
	err = json.Unmarshal([]byte(connectionStr4), &connection4)
	if err != nil {
		panic(err)
	}

	dummyCatalog.dataDetails["db2"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "db2",
			Geography: &geo4,
			Tags:      &tags,
		},
		Details: catalogmodels.Details{
			Connection: connection4,
			DataFormat: &dataformat4},
		Credentials: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	}

	// dummyCatalog.dataDetails["db2"] = pb.CatalogDatasetInfo{
	// 	DatasetId: "db2",
	// 	Details: &pb.DatasetDetails{
	// 		Name:       "yyy",
	// 		DataFormat: "table",
	// 		Geo:        "theshire",
	// 		DataStore: &pb.DataStore{
	// 			Type: pb.DataStore_DB2,
	// 			Name: "db2",
	// 			Db2: &pb.Db2DataStore{
	// 				Database: "BLUDB",
	// 				Table:    "NQD60833.SMALL",
	// 				Url:      "dashdb-txn-sbox-yp-lon02-02.services.eu-gb.bluemix.net",
	// 				Port:     "50000",
	// 				Ssl:      "false",
	// 			},
	// 		},
	// 		CredentialsInfo: &pb.CredentialsInfo{
	// 			VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	// 		},
	// 		Metadata: &pb.DatasetMetadata{},
	// 	},
	// }

	geo5 := "theshire"
	dataformat5 := "json"
	var connection5 catalogmodels.Connection
	var connectionStr5 = `{
		"name": "Kafka",
		"kafka": {
			"TopicName":             "topic",
			"SecurityProtocol":      "SASL_SSL",
			"SaslMechanism":         "SCRAM-SHA-512",
			"SslTruststore":         "xyz123",
			"SslTruststorePassword": "passwd",
			"SchemaRegistry":        "kafka-registry",
			"BootstrapServers":      "http://kafka-servers",
			"KeyDeserializer":       "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer",
			"ValueDeserializer":     "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer"
			}
		}`
	err = json.Unmarshal([]byte(connectionStr5), &connection5)
	if err != nil {
		panic(err)
	}

	dummyCatalog.dataDetails["kafka"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "Cars",
			Geography: &geo5,
			Tags:      &tags,
		},
		Details: catalogmodels.Details{
			Connection: connection5,
			DataFormat: &dataformat5},
		Credentials: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	}

	// dummyCatalog.dataDetails["kafka"] = pb.CatalogDatasetInfo{
	// 	DatasetId: "kafka",
	// 	Details: &pb.DatasetDetails{
	// 		Name:       "Cars",
	// 		DataFormat: "json",
	// 		Geo:        "theshire",
	// 		DataStore: &pb.DataStore{
	// 			Type: pb.DataStore_KAFKA,
	// 			Name: "kafka",
	// 			Kafka: &pb.KafkaDataStore{
	// 				TopicName:             "topic",
	// 				SecurityProtocol:      "SASL_SSL",
	// 				SaslMechanism:         "SCRAM-SHA-512",
	// 				SslTruststore:         "xyz123",
	// 				SslTruststorePassword: "passwd",
	// 				SchemaRegistry:        "kafka-registry",
	// 				BootstrapServers:      "http://kafka-servers",
	// 				KeyDeserializer:       "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer",
	// 				ValueDeserializer:     "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer",
	// 			},
	// 		},
	// 		CredentialsInfo: &pb.CredentialsInfo{
	// 			VaultSecretPath: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	// 		},
	// 		Metadata: &pb.DatasetMetadata{},
	// 	},
	// }

	geo6 := "theshire"
	dataformat6 := "csv"
	var connection6 catalogmodels.Connection
	var connectionStr6 = `{
		"name": "local"
		}`
	err = json.Unmarshal([]byte(connectionStr6), &connection6)
	if err != nil {
		panic(err)
	}

	dummyCatalog.dataDetails["local"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "local file",
			Geography: &geo6,
			Tags:      &tags,
		},
		Details: catalogmodels.Details{
			Connection: connection6,
			DataFormat: &dataformat6},
		Credentials: "/v1/kubernetes-secrets/creds-secret-name?namespace=fybrik-system",
	}

	// dummyCatalog.dataDetails["local"] = pb.CatalogDatasetInfo{
	// 	DatasetId: "local",
	// 	Details: &pb.DatasetDetails{
	// 		Name:       "local file",
	// 		DataFormat: "csv",
	// 		Geo:        "theshire",
	// 		DataStore: &pb.DataStore{
	// 			Type: pb.DataStore_LOCAL,
	// 			Name: "file.csv",
	// 		},
	// 		Metadata: &pb.DatasetMetadata{DatasetTags: []string{"PI"}},
	// 	},
	// }

	return &dummyCatalog
}
