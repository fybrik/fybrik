// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"fybrik.io/fybrik/pkg/model/datacatalog"
)

func main() {
	request := datacatalog.CreateAssetRequest{
		DestinationCatalogID: "test",
	}

	postBody, _ := json.Marshal(request)
	responseBody := bytes.NewBuffer(postBody)

	fmt.Println("Making a HTTP Request")
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://katalog-connector:80/createAssetInfo", responseBody)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	var responseObject datacatalog.CreateAssetResponse
	err = json.Unmarshal(bodyBytes, &responseObject)
	if err != nil {
		fmt.Print(err.Error())
	} else {
		fmt.Printf("CreateAssetResponse received: %+v\n", responseObject)
	}
}
