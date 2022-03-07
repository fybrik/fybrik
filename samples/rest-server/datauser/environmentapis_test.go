// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func SkipOnClosedSocket(address string, t *testing.T) {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		t.Skip("Skipping test as server is not running...")
	}
	if conn != nil {
		defer conn.Close()
	}
}

func TestEnvironmentAPIs(t *testing.T) {
	envserverurl := "http://localhost:8080/v1/env/datauserenv"

	SkipOnClosedSocket("localhost:8080", t)

	sysMap := make(map[string][]string)
	sysMap["Egeria"] = []string{"username"}
	expectedObj := EnvironmentInfo{
		Namespace:       "default",
		Geography:       "thegreendragon",
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
	respData, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	respString := string(respData)

	err = json.Unmarshal([]byte(respString), &resultObj)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resultObj, expectedObj, "Wrong response")
}
