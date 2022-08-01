// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mockup

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
)

type DataCatalogDummy struct {
	dataDetails map[string]datacatalog.GetAssetResponse
}

func (d *DataCatalogDummy) GetAssetInfo(in *datacatalog.GetAssetRequest, creds string) (*datacatalog.GetAssetResponse, error) {
	datasetID := string(in.AssetID)
	log.Printf("MockDataCatalog.GetDatasetInfo called with DataSetID " + datasetID)

	splittedID := strings.SplitN(datasetID, "/", 2)
	if len(splittedID) != 2 {
		panic(fmt.Sprintf("Invalid dataset ID for mock: %s", datasetID))
	}

	catalogID := splittedID[0]

	dataDetails, found := d.dataDetails[catalogID]
	if found {
		log.Printf("GetAssetInfo in DataCatalogDummy returns:")
		responseBytes, errJSON := json.MarshalIndent(&dataDetails, "", "\t")
		if errJSON != nil {
			return nil, fmt.Errorf("error in GetAssetInfo in DataCatalogDummy: %v", errJSON)
		}
		log.Print(string(responseBytes))
		return &dataDetails, nil
	}

	return nil, errors.New("the asset does not exist")
}

func (d *DataCatalogDummy) CreateAsset(in *datacatalog.CreateAssetRequest, creds string) (*datacatalog.CreateAssetResponse, error) {
	// TODO: will be provided a proper implementation once the implementation of CreateAsset in katalog-connector
	// is completed in a future PR. Till then a dummy implementation is provided.
	return &datacatalog.CreateAssetResponse{AssetID: "testAssetID"}, nil
}

func (d *DataCatalogDummy) DeleteAsset(in *datacatalog.DeleteAssetRequest, creds string) (*datacatalog.DeleteAssetResponse, error) {
	// TODO: will be provided a proper implementation once the implementation of DeleteAsset in katalog-connector
	// is completed in a future PR. Till then a dummy implementation is provided.
	return &datacatalog.DeleteAssetResponse{Status: "DeleteAssetInfo not implemented in DataCatalogDummy"}, nil
}

func (m *DataCatalogDummy) UpdateAsset(in *datacatalog.UpdateAssetRequest, creds string) (*datacatalog.UpdateAssetResponse, error) {
	// TODO: will be provided a proper implementation once the implementation of UpdateAsset in katalog-connector
	// is completed in a future PR. Till then a dummy implementation is provided.
	return &datacatalog.UpdateAssetResponse{Status: "UpdateAsset not implemented in DataCatalogDummy"}, nil
}

func (d *DataCatalogDummy) Close() error {
	return nil
}

//nolint:funlen
func NewTestCatalog() *DataCatalogDummy {
	dummyCatalog := DataCatalogDummy{
		dataDetails: make(map[string]datacatalog.GetAssetResponse),
	}

	tags := taxonomy.Tags{}
	tags.Items = map[string]interface{}{"PI": true}

	geo := "theshire" //nolint:goconst
	geoExternal := "neverland"

	// TODO(roee88): some of these are also defined in ifdetails.go
	var csvFormat taxonomy.DataFormat = "csv"
	var parquetFormat taxonomy.DataFormat = "parquet"
	var jsonFormat taxonomy.DataFormat = "json"
	dummyResourceName := "xxx"
	dummyCredentials := "dummy"

	s3Connection := taxonomy.Connection{
		Name: S3,
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				string(S3): map[string]interface{}{
					// TODO(roee88): why are real endpoints used?
					"endpoint":   "s3.eu-gb.cloud-object-storage.appdomain.cloud",
					"bucket":     "fybrik-test-bucket",
					"object_key": "small.csv",
				},
			},
		},
	}

	db2Connection := taxonomy.Connection{
		Name: JdbcDB2,
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				string(JdbcDB2): map[string]interface{}{
					"database": "test-db",
					"table":    "test-table",
					"url":      "dashdb-txn-sbox-yp-lon02-02.services.eu-gb.bluemix.net",
					"port":     "5000",  // TODO(roee88): should this be defined in the example taxonomy as integer?
					"ssl":      "false", // TODO(roee88): should this be defined in the example taxonomy as boolean?
				},
			},
		},
	}
	kafkaDeserializer := "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer"
	sslTruststore := "xyz123"
	kafkaConnection := taxonomy.Connection{
		Name: Kafka,
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				string(Kafka): map[string]interface{}{
					"topic_name":              "topic",
					"security_protocol":       "SASL_SSL",
					"sasl_mechanism":          "SCRAM-SHA-512",
					"ssl_truststore":          sslTruststore,
					"ssl_truststore_password": sslTruststore,
					"schema_registry":         "kafka-registry",
					"bootstrap_servers":       "http://kafka-servers",
					"key_deserializer":        kafkaDeserializer,
					"value_deserializer":      kafkaDeserializer,
				},
			},
		},
	}

	dummyCatalog.dataDetails["s3-external"] = datacatalog.GetAssetResponse{
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name:      dummyResourceName,
			Geography: geoExternal,
			Tags:      &tags,
		},
		Credentials: dummyCredentials,
		Details: datacatalog.ResourceDetails{
			Connection: s3Connection,
			DataFormat: csvFormat,
		},
	}

	dummyCatalog.dataDetails[string(S3)] = datacatalog.GetAssetResponse{
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name:      dummyResourceName,
			Geography: geo,
			Tags:      &tags,
		},
		Credentials: dummyCredentials,
		Details: datacatalog.ResourceDetails{
			Connection: s3Connection,
			DataFormat: parquetFormat,
		},
	}

	dummyCatalog.dataDetails["s3-csv"] = datacatalog.GetAssetResponse{
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name:      dummyResourceName,
			Geography: geo,
			Tags:      &tags,
		},
		Credentials: dummyCredentials,
		Details: datacatalog.ResourceDetails{
			Connection: s3Connection,
			DataFormat: csvFormat,
		},
	}

	dummyCatalog.dataDetails[string(JdbcDB2)] = datacatalog.GetAssetResponse{
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name:      dummyResourceName,
			Geography: geo,
			Tags:      &tags,
		},
		Credentials: dummyCredentials,
		Details: datacatalog.ResourceDetails{
			Connection: db2Connection,
		},
	}

	dummyCatalog.dataDetails[string(Kafka)] = datacatalog.GetAssetResponse{
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name:      dummyResourceName,
			Geography: geo,
			Tags:      &tags,
		},
		Credentials: dummyCredentials,
		Details: datacatalog.ResourceDetails{
			Connection: kafkaConnection,
			DataFormat: jsonFormat,
		},
	}
	return &dummyCatalog
}
