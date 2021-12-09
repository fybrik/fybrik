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
	dummyCatalog.dataDetails["s3-external"] = catalogmodels.DataCatalogResponse{
		ResourceMetadata: catalogmodels.Resource{
			Name:      "xxx",
			Geography: &geo_external,
			Tags:      &tags,
		},
		Credentials: "dummy",
		Details: catalogmodels.Details{
			Connection: catalogmodels.Connection{
				Name: "s3",
				AdditionalProperties: map[string]interface{}{
					"endpoint":  "s3.cloud-object-storage",
					"bucket":    "test-bucket",
					"objectKey": "test.csv",
				},
			},
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
			Connection: catalogmodels.Connection{
				Name: "s3",
				AdditionalProperties: map[string]interface{}{
					"endpoint":  "s3.cloud-object-storage",
					"bucket":    "test-bucket",
					"objectKey": "test.parq",
				},
			},
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
			Connection: catalogmodels.Connection{
				Name: "s3",
				AdditionalProperties: map[string]interface{}{
					"endpoint":  "s3.cloud-object-storage",
					"bucket":    "test-bucket",
					"objectKey": "test.csv",
				},
			},
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
			Connection: catalogmodels.Connection{
				Name: "jdbc-db2",
				AdditionalProperties: map[string]interface{}{
					"database": "BLUDB",
					"table":    "NQD60833.SMALL",
					"url":      "dashdb-txn-sbox-yp-lon02-02.services.eu-gb.bluemix.net",
					"port":     "50000",
					"ssl":      "false",
				},
			},
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
			Connection: catalogmodels.Connection{
				Name: "kafka",
				AdditionalProperties: map[string]interface{}{
					"topicName":             "topic",
					"securityProtocol":      "SASL_SSL",
					"saslMechanism":         "SCRAM-SHA-512",
					"sslTruststore":         "xyz123",
					"sslTruststorePassword": "passwd",
					"schemaRegistry":        "kafka-registry",
					"bootstrapServers":      "http://kafka-servers",
					"keyDeserializer":       "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer",
					"valueDeserializer":     "io.confluent.kafka.serializers.json.KafkaJsonSchemaDeserializer",
				},
			},
			DataFormat: &json_format,
		},
	}
	return &dummyCatalog
}
