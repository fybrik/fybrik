// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentAPIs(t *testing.T) {
	SkipOnClosedSocket("localhost:8080", t)
	expectedObj := EnvironmentInfo{
		Namespace: "default",
		Geography: os.Getenv("GEOGRAPHY"),
	}
	// Call the REST API to get the environment information
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:8080/v1/env/datauserenv", http.NoBody)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
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
