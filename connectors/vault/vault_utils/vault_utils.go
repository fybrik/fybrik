// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vaultutils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const (
	VaultAddressKey       string = "USER_VAULT_ADDRESS"
	VaultSecretKey        string = "USER_VAULT_TOKEN"
	VaultTimeoutKey       string = "USER_VAULT_TIMEOUT"
	VaultPathKey          string = "USER_VAULT_PATH"
	VaultConnectorPortKey string = "PORT_VAULT_CONNECTOR"
	DefaultTimeout        string = "180"
	DefaultPort           string = "50083" // synced with vault_connector.yaml
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
		return nil, errors.Wrap(err, "error creating vault client")
	}

	// Get the vault token stored in config
	if token == "" {
		return nil, errors.New("cannot authenticate with vault: no vault token found")
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
		return "", errors.New("no logical client received when retrieving credentials from vault")
	}

	// logicalClient does not work with paths that start with /v1/ so we need to remove the prefix
	if strings.HasPrefix(vaultPath, "/v1/") {
		vaultPath = vaultPath[3:]
	}

	data, err := logicalClient.Read(vaultPath)
	if err != nil {
		return "", errors.Wrapf(err, "error reading credentials from vault for %s", vaultPath)
	}

	if data == nil || data.Data == nil {
		return "", fmt.Errorf("no data received for credentials from vault for %s", vaultPath)
	}

	b, jsonErr := json.Marshal(data.Data)
	if jsonErr != nil {
		return "", errors.Wrapf(err, "error marshaling credentials to json for %s", vaultPath)
	}

	return string(b), nil
}

// AddToVault adds crededentialsMap to vault at the path given by innerVaultPath
func (vlt *VaultConnection) AddToVault(innerVaultPath string, credentialsMap map[string]interface{}) (string, error) {
	vaultDatasetPath := innerVaultPath

	// Add credentials to vault, and return vaultPath where they are stored
	logicalClient := vlt.Client.Logical()
	if logicalClient == nil {
		return vaultDatasetPath, errors.New("no logical client received when adding data set credentials to vault")
	}

	log.Printf("vaultDatasetPath in AddToVault: %s\n", vaultDatasetPath)
	_, err := logicalClient.Write(vaultDatasetPath, credentialsMap)
	if err != nil {
		return vaultDatasetPath, errors.Wrapf(err, "error adding credentials to vault to %s", vaultDatasetPath)
	}
	return vaultDatasetPath, nil
}
