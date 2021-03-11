// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentAPIs(t *testing.T) {
	envserverurl := "http://localhost:8080/v1/env/datauserenv"

	sysMap := make(map[string][]string)
	sysMap["Egeria"] = []string{"username"}
	expectedObj := EnvironmentInfo{
		Namespace:       "default",
		Geography:       "US-cluster",
		Systems:         sysMap,
		DataSetIDFormat: "{\"ServerName\":\"---\",\"AssetGuid\":\"---\"}",
	}
	// Call the REST API to get the environment information
	resp, err := http.Get(envserverurl)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "expected status OK")

	// Check that the results are what we expected
	// Convert result from json to go struct
	resultObj := EnvironmentInfo{}
	respData, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	respString := string(respData)

	err = json.Unmarshal([]byte(respString), &resultObj)
	assert.Nil(t, err)

	assert.Equal(t, resultObj, expectedObj, "Wrong response")

	defer resp.Body.Close()
}
