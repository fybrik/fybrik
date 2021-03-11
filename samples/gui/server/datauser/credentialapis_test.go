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
	credserverurl = "http://localhost:8080/v1/creds/usercredentials"
	cred1         = "{\"SecretName\": \"notebook\",\"System\": \"Egeria\",\"Credentials\": {\"username\": \"user1\"}}"
	cred2         = "{\"SecretName\": \"notebook\",\"System\": \"OPA\", \"Credentials\": {\"username\": \"user2\"}}"
	name          = "notebook"
)

func storeCredentials(t *testing.T, cred string) {
	body := strings.NewReader(cred)
	req, err := http.NewRequest("POST", credserverurl, body)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusCreated, "Failed to store credentials")
	defer resp.Body.Close()
}

func readCredentials(t *testing.T, path string) {
	url := credserverurl + "/" + path
	resp, err := http.Get(url)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Failed to read credentials")
	defer resp.Body.Close()
}

func deleteCredentials(t *testing.T, path string) {
	url := credserverurl + "/" + path
	req, err := http.NewRequest("DELETE", url, nil)
	assert.Nil(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "Failed to delete credentials")

	defer resp.Body.Close()
}

func TestCredentialAPIs(t *testing.T) {
	storeCredentials(t, cred1)
	storeCredentials(t, cred2)
	readCredentials(t, name)
	deleteCredentials(t, name)
}
