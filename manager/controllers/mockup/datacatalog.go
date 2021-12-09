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
	geo := "theshire"
	geo_external := "neverland"
	csv_format := "csv"
	parquet_format := "parquet"
	db2_format := "table"
	json_format := "json"

	s3_connection := catalogmodels.Connection{}
	s3_map := make(map[string]interface{})
	s3_map["name"] = "s3"
	s3_map["endpoint"] = "s3.cloud-object-storage"
	s3_map["bucket"] = "test-bucket"
	s3_map["objectKey"] = "test"
	bytes, _ := json.MarshalIndent(s3_map, "", "\t")
	_ = json.Unmarshal(bytes, &s3_connection)

	db2_connection := catalogmodels.Connection{}
	db2_map := make(map[string]interface{})
	db2_map["name"] = "jdbc-db2"
	db2_map["database"] = "test-db"
	db2_map["table"] = "test-table"
	db2_map["url"] = "dashdb-txn-sbox-yp-lon02-02.services.eu-gb.bluemix.net"
	db2_map["port"] = "5000"
	db2_map["ssl"] = "false"
	bytes, _ = json.MarshalIndent(db2_map, "", "\t")
	_ = json.Unmarshal(bytes, &db2_connection)

	kafka_connection := catalogmodels.Connection{}
	kafka_map := make(map[string]interface{})
	kafka_map["name"] = "kafka"
	kafka_map["topicName"] = "topic"
	kafka_map["securityProtocol"] = "SASL_SSL"
	kafka_map["saslMechanism"] = "SCRAM-SHA-512"
	kafka_map["sslTruststore"] = "xyz123"
	kafka_map["sslTruststorePassword"] = "passwd"
	kafka_map["schemaRegistry"] = "kafka-registry"
	kafka_map["bootstrapServers"] = "http://kafka-servers"
	kafka_map["keyDeserializer"] = "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer"
	kafka_map["valueDeserializer"] = "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer"

	bytes, _ = json.MarshalIndent(kafka_map, "", "\t")
	_ = json.Unmarshal(bytes, &kafka_connection)

	dummyCatalog.dataDetails["s3-external"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "xxx",
			Geography: &geo_external,
			Tags:      &tags,
		},
		Credentials: "dummy",
		Details: catalogmodels.Details{
			Connection: s3_connection,
			DataFormat: &csv_format,
		},
	}

	dummyCatalog.dataDetails["s3"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "xxx",
			Geography: &geo,
			Tags:      &tags,
		},
		Credentials: "dummy",
		Details: catalogmodels.Details{
			Connection: s3_connection,
			DataFormat: &parquet_format,
		},
	}

	dummyCatalog.dataDetails["s3-csv"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "xxx",
			Geography: &geo,
			Tags:      &tags,
		},
		Credentials: "dummy",
		Details: catalogmodels.Details{
			Connection: s3_connection,
			DataFormat: &csv_format,
		},
	}

	dummyCatalog.dataDetails["db2"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "xxx",
			Geography: &geo,
			Tags:      &tags,
		},
		Credentials: "dummy",
		Details: catalogmodels.Details{
			Connection: db2_connection,
			DataFormat: &db2_format,
		},
	}

	dummyCatalog.dataDetails["kafka"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "xxx",
			Geography: &geo,
			Tags:      &tags,
		},
		Credentials: "dummy",
		Details: catalogmodels.Details{
			Connection: kafka_connection,
			DataFormat: &json_format,
		},
	}
	return &dummyCatalog
}
