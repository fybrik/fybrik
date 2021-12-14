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
	geoExternal := "neverland"
	csvFormat := "csv"
	parquetFormat := "parquet"
	db2Format := "table"
	jsonFormat := "json"

	s3Connection := catalogmodels.Connection{}
	s3Config := make(map[string]interface{})
	s3Map := make(map[string]interface{})
	s3Map["endpoint"] = "s3.cloud-object-storage"
	s3Map["bucket"] = "test-bucket"
	s3Map["objectKey"] = "test"
	s3Config["name"] = "s3"
	s3Config["s3"] = s3Map

	bytes, _ := json.MarshalIndent(s3Config, "", "\t")
	_ = json.Unmarshal(bytes, &s3Connection)

	db2Connection := catalogmodels.Connection{}
	db2Map := make(map[string]interface{})
	db2Map["database"] = "test-db"
	db2Map["table"] = "test-table"
	db2Map["url"] = "dashdb-txn-sbox-yp-lon02-02.services.eu-gb.bluemix.net"
	db2Map["port"] = "5000"
	db2Map["ssl"] = "false"
	db2Config := make(map[string]interface{})
	db2Config["name"] = "jdbc-db2"
	db2Config["jdbc-db2"] = db2Map
	bytes, _ = json.MarshalIndent(db2Config, "", "\t")
	_ = json.Unmarshal(bytes, &db2Connection)

	kafkaConnection := catalogmodels.Connection{}
	kafkaMap := make(map[string]interface{})
	kafkaMap["topic_name"] = "topic"
	kafkaMap["security_protocol"] = "SASL_SSL"
	kafkaMap["sasl_mechanism"] = "SCRAM-SHA-512"
	kafkaMap["ssl_truststore"] = "xyz123"
	kafkaMap["ssl_truststore_password"] = "passwd"
	kafkaMap["schema_registry"] = "kafka-registry"
	kafkaMap["bootstrap_servers"] = "http://kafka-servers"
	kafkaMap["key_deserializer"] = "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer"
	kafkaMap["value_deserializer"] = "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer"
	kafkaConfig := make(map[string]interface{})
	kafkaConfig["name"] = "kafka"
	kafkaConfig["kafka"] = kafkaMap
	bytes, _ = json.MarshalIndent(kafkaConfig, "", "\t")
	_ = json.Unmarshal(bytes, &kafkaConnection)

	dummyCatalog.dataDetails["s3-external"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "xxx",
			Geography: &geoExternal,
			Tags:      &tags,
		},
		Credentials: "dummy",
		Details: catalogmodels.Details{
			Connection: s3Connection,
			DataFormat: &csvFormat,
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
			Connection: s3Connection,
			DataFormat: &parquetFormat,
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
			Connection: s3Connection,
			DataFormat: &csvFormat,
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
			Connection: db2Connection,
			DataFormat: &db2Format,
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
			Connection: kafkaConnection,
			DataFormat: &jsonFormat,
		},
	}
	return &dummyCatalog
}
