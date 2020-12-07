// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vaultutils

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
	DefaultTimeout        string = "180"
	DefaultPort           string = "50083" //synched with vault_connector.yaml
)

func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Env Variable %v not defined", key)
	}
	if key != VaultSecretKey {
		log.Printf("Env. variable extracted: %s - %s\n", key, value)
	}
	return value
}

func GetEnvWithDefault(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Env. variable not found, use default value: %s - %s\n", key, defaultValue)
		return defaultValue
	}
	log.Printf("Env. variable extracted: %s - %s\n", key, value)
	return value
}

type VaultConfig struct {
	Token   string
	Address string
}

type VaultConnection struct {
	Config VaultConfig
	Client *api.Client
}

func CreateVaultConnection() VaultConnection {
	token := GetEnv(VaultSecretKey)
	address := GetEnv(VaultAddressKey)
	config := VaultConfig{
		Token:   token,
		Address: address,
	}

	connection := VaultConnection{
		Config: config,
	}

	client, _ := connection.InitVault()
	connection.Client = client

	return connection
}

func (vlt *VaultConnection) InitVault() (*api.Client, error) {
	vaultAddress := vlt.Config.Address
	token := vlt.Config.Token

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
func (vlt *VaultConnection) GetFromVault(vaultPathKey string, innerVaultPath string) (string, error) {
	vaultPath := vaultPathKey + "/" + innerVaultPath

	logicalClient := vlt.Client.Logical()
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

// AddToVault adds crededentialsMap to vault at the path given by innerVaultPath
func (vlt *VaultConnection) AddToVault(innerVaultPath string, credentialsMap map[string]interface{}) (string, error) {
	vaultDatasetPath := innerVaultPath

	// Add credentials to vault, and return vaultPath where they are stored
	logicalClient := vlt.Client.Logical()
	if logicalClient == nil {
		msg := "No logical client received when adding data set credentials to vault"
		return vaultDatasetPath, errors.New(msg)
	}

	log.Printf("vaultDatasetPath in AddToVault: %s\n", vaultDatasetPath)
	_, err := logicalClient.Write(vaultDatasetPath, credentialsMap)
	if err != nil {
		msg := "Error adding credentials to vault to " + vaultDatasetPath + ":" + err.Error()
		return vaultDatasetPath, errors.New(msg)
	}
	return vaultDatasetPath, nil
}
