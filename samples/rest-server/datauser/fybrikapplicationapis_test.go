// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	dmaserverurl = "http://localhost:8080/v1/dma/fybrikapplication"
	dm1          = "{\"apiVersion\": \"app.fybrik.io/v12\",\"kind\": \"FybrikApplication\",\"metadata\": {\"name\": \"unittest-read\"}," +
		"\"spec\": {\"selector\": {\"workloadSelector\": {\"matchLabels\":{\"app\": \"notebook\"}}}," +
		"\"appInfo\": {\"intent\": \"fraud-detection\"}, \"data\": [{\"dataSetID\": \"123\",\"requirements\": " +
		"{ \"interface\": {\"protocol\": \"s3\",\"dataformat\": \"parquet\"}}}]}}"
	dm1name = "unittest-read"
	dm2     = "{\"apiVersion\": \"app.fybrik.io/v12\",\"kind\": \"FybrikApplication\",\"metadata\": {\"name\": \"unittest-copy\"}," +
		"\"spec\": {\"selector\": {\"workloadSelector\": {}}," +
		"\"appInfo\": {}, \"data\": [{\"dataSetID\": \"456\", \"flow\": \"copy\", \"requirements\": " +
		"{\"flowParams\": {\"catalog\": \"enterprise\"}, \"interface\": " +
		"{\"protocol\": \"s3\",\"dataformat\": \"parquet\"}}}]}}"
	dm2name = "unittest-copy"
)

func createApplication(t *testing.T, obj, name string) {
	body := strings.NewReader(obj)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, dmaserverurl, body)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusCreated, "Failed to create application "+name)
}

func listApplications(t *testing.T) {
	url := dmaserverurl
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Failed to list applications")
}

func getApplication(t *testing.T, name string) {
	url := dmaserverurl + "/" + name
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Failed to get application "+name)
}

func deleteApplication(t *testing.T, name string) {
	url := dmaserverurl + "/" + name
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, url, http.NoBody)
	assert.Nil(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Failed to delete application "+name)
}

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

func TestApplicationAPIs(t *testing.T) {
	SkipOnClosedSocket("localhost:8080", t)
	createApplication(t, dm1, dm1name)
	createApplication(t, dm2, dm2name)
	listApplications(t)
	getApplication(t, dm1name)
	getApplication(t, dm2name)
	deleteApplication(t, dm1name)
	deleteApplication(t, dm2name)
}
