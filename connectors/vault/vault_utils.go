// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

const (
	VaultAddressKey       string = "USER_VAULT_ADDRESS"
	VaultSecretKey        string = "USER_VAULT_TOKEN"
	VaultTimeoutKey       string = "USER_VAULT_TIMEOUT"
	VaultPathKey          string = "USER_VAULT_PATH"
	VaultConnectorPortKey string = "PORT_VAULT_CONNECTOR"
)

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Env Variable %v not defined", key)
	}
	log.Printf("Env. variable extracted: %s - %s\n", key, value)
	return value
}

func getEnvWithDefault(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Env. variable not found, use default value: %s - %s\n", key, defaultValue)
		return defaultValue
	}
	log.Printf("Env. variable extracted: %s - %s\n", key, value)
	return value
}

type VaultConfig struct {
	token   string
	address string
}

type VaultConnection struct {
	config VaultConfig
	client *api.Client
}

func CreateVaultConnection() VaultConnection {
	token := getEnv(VaultSecretKey)
	address := getEnv(VaultAddressKey)
	config := VaultConfig{
		token:   token,
		address: address,
	}

	connection := VaultConnection{
		config: config,
	}

	client, _ := connection.InitVault()
	connection.client = client

	return connection
}

func (vlt *VaultConnection) InitVault() (*api.Client, error) {
	vaultAddress := vlt.config.address
	token := vlt.config.token

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	conf := &api.Config{
		Address:    vaultAddress,
		HttpClient: httpClient,
	}

	client, err := api.NewClient(conf)
	if err != nil {
		msg := "Error creating vault client: " + err.Error()
		return nil, errors.New(msg)
	}

	// Get the vault token stored in config
	if token == "" {
		msg := "No vault token found.  Cannot authenticate with vault."
		return nil, errors.New(msg)
	}

	client.SetToken(token)
	log.Println("Token set successfully")

	return client, nil
}

// GetFromVault returns the credentials from vault as json
func (vlt *VaultConnection) GetFromVault(innerVaultPath string) (string, error) {
	vaultPath := getEnv(VaultPathKey) + "/" + innerVaultPath

	logicalClient := vlt.client.Logical()
	if logicalClient == nil {
		msg := "No logical client received when retrieving credentials from vault"
		return "", errors.New(msg)
	}

	// logicalClient does not work with paths that start with /v1/ so we need to remove the prefix
	if strings.HasPrefix(vaultPath, "/v1/") {
		vaultPath = vaultPath[3:]
	}

	data, err := logicalClient.Read(vaultPath)
	if err != nil {
		msg := "Error reading credentials from vault for " + vaultPath + ":" + err.Error()
		return "", errors.New(msg)
	}

	if data == nil || data.Data == nil {
		msg := "No data received for credentials from vault for " + vaultPath
		return "", errors.New(msg)
	}

	b, jsonErr := json.Marshal(data.Data)
	if jsonErr != nil {
		msg := "Error marshaling credentials to json for " + vaultPath + ":" + jsonErr.Error()
		return "", errors.New(msg)
	}

	return string(b), nil
}
