// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
)

func main() {
	// the below logic is inspired from https://levelup.gitconnected.com/consuming-a-rest-api-using-golang-b323602ba9d8

	// ResourceMetadata   ResourceMetadata `json:"resourceMetadata"`
	// Details            ResourceDetails  `json:"details"`
	// // This has the vault plugin path where the data credentials will be stored as kubernetes secrets
	// // This value is assumed to be known to the catalog connector.
	// Credentials string `json:"credentials"`

	request := datacatalog.CreateAssetRequest{
		DestinationCatalogID: "testcatalogid",
		Credentials:          "http://fybrik-system:8200/v1/kubernetes-secrets/wkc-creds?namespace=cp4d",
		Details: datacatalog.ResourceDetails{
			Connection: taxonomy.Connection{
				Name: "db2",
			},
		},
		ResourceMetadata: datacatalog.ResourceMetadata{
			Name:      "demoAsset",
			Owner:     "Alice",
			Geography: "us-south",
			Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
				"finance": true,
			}}},
			Columns: []datacatalog.ResourceColumn{
				{
					Name: "c1",
					Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
						"PII": true,
					}}},
				},
			},
		},
		// ResourceMetadata: datacatalog.ResourceMetadata{
		// 	Name: "assetName",
		// 	Columns: []datacatalog.ResourceColumn{
		// 		{
		// 			Name: "nameDest",
		// 			Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
		// 				"PII": true,
		// 			}}},
		// 		},
		// 		{
		// 			Name: "nameOrig",
		// 			Tags: &taxonomy.Tags{Properties: serde.Properties{Items: map[string]interface{}{
		// 				"SPI": true,
		// 			}}},
		// 		},
		// 	},
		// },
	}

	postBody, _ := json.Marshal(request)
	responseBody := bytes.NewBuffer(postBody)

	fmt.Println("Making a HTTP Request")
	fmt.Println("responseBody:", responseBody)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://wkc-connector:8080/createAssetInfo", responseBody)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Request-Datacatalog-Write-Cred", "http://fybrik-system:8200/v1/kubernetes-secrets/wkc-creds?namespace=cp4d")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print("here1")
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Print("here2")
		fmt.Print(err.Error())
	}
	var responseObject datacatalog.CreateAssetResponse
	err = json.Unmarshal(bodyBytes, &responseObject)
	if err != nil {
		fmt.Print("here3")
		fmt.Print(err.Error())
	} else {
		fmt.Print("here4")
		fmt.Printf("CreateAssetResponse received: %+v\n", responseObject)
	}
}
