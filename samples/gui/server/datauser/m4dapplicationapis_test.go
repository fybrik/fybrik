// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dmaserverurl = "http://localhost:8080/v1/dma/m4dapplication"
	dm1          = "{\"apiVersion\": \"app.m4d.ibm.com/v1alpha1\",\"kind\": \"M4DApplication\",\"metadata\": {\"name\": \"unittest-read\"},\"spec\": {\"selector\": {\"workloadSelector\": {\"matchLabels\":{\"app\": \"notebook\"}}},\"appInfo\": {\"intent\": \"fraud-detection\"}, \"data\": [{\"dataSetID\": \"123\",\"requirements\": { \"interface\": {\"protocol\": \"s3\",\"dataformat\": \"parquet\"}}}]}}"
	dm1name      = "unittest-read"
	dm2          = "{\"apiVersion\": \"app.m4d.ibm.com/v1alpha1\",\"kind\": \"M4DApplication\",\"metadata\": {\"name\": \"unittest-copy\"},\"spec\": {\"selector\": {\"workloadSelector\": {}},\"appInfo\": {\"intent\": \"copy data\"}, \"data\": [{\"dataSetID\": \"456\",\"requirements\": {\"copy\": {\"required\": true,\"catalog\": {\"catalogID\": \"enterprise\"}}, \"interface\": {\"protocol\": \"s3\",\"dataformat\": \"parquet\"}}}]}}"
	dm2name      = "unittest-copy"
)

func createApplication(t *testing.T, obj string, name string) {
	body := strings.NewReader(obj)
	req, err := http.NewRequest("POST", dmaserverurl, body)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusCreated, "Failed to create application "+name)
	defer resp.Body.Close()
}

func listApplications(t *testing.T) {
	url := dmaserverurl
	resp, err := http.Get(url)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Failed to list applications")
	defer resp.Body.Close()
}

func getApplication(t *testing.T, name string) {
	url := dmaserverurl + "/" + name
	resp, err := http.Get(url)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Failed to get application "+name)
	defer resp.Body.Close()
}

func deleteApplication(t *testing.T, name string) {
	url := dmaserverurl + "/" + name
	req, err := http.NewRequest("DELETE", url, nil)
	assert.Nil(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Failed to delete application "+name)

	defer resp.Body.Close()
}

func TestApplicationAPIs(t *testing.T) {
	createApplication(t, dm1, dm1name)
	createApplication(t, dm2, dm2name)
	listApplications(t)
	getApplication(t, dm1name)
	getApplication(t, dm2name)
	deleteApplication(t, dm1name)
	deleteApplication(t, dm2name)
}
