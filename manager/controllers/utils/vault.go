// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// LinkVaultPolicyToIdentity registers a policy for a given identity or role, meaning that when a person or service
// of that identity logs into vault and tries to read or write a secret the provided policy
// will determine whether that is allowed or not.
func LinkVaultPolicyToIdentity(identity string, policyName string, vaultClient *api.Client) error {
	if RunWithoutVaultHook() {
		return nil
	}

	auth, _ := GetVaultAuthService()
	identityPath := GetIdentitiesVaultAuthPath() + "/" + identity

	logicalClient := vaultClient.Logical()
	if logicalClient == nil {
		msg := "No logical client received when linking policy " + policyName + " to idenity " + identity
		return errors.New(msg)
	}

	params := map[string]interface{}{
		"user_claim":                       "sub",
		"role_type":                        auth,
		"bound_service_account_names":      "secret-provider",
		"bound_service_account_namespaces": GetSystemNamespace(),
		"policies":                         policyName,
		"ttl":                              GetVaultAuthTTL(),
	}

	_, err := logicalClient.Write(identityPath, params)
	if err != nil {
		msg := "Error linking policy " + policyName + " to identity " + identity + ":" + err.Error()
		return errors.New(msg)
	}

	return nil
}

// RemoveVaultPolicyFromIdentity removes the policy from the authentication identity with which it is associated, meaning
// this policy will no longer be invoked when a person or service authenticates with this identity.
// TODO - Implement this
func RemoveVaultPolicyFromIdentity(identity string, policyName string, vaultClient *api.Client) error {
	if RunWithoutVaultHook() {
		return nil
	}

	identityPath := GetIdentitiesVaultAuthPath() + "/" + identity

	logicalClient := vaultClient.Logical()
	if logicalClient == nil {
		msg := "No logical client received when deleting policy " + policyName + " from idenity " + identity
		return errors.New(msg)
	}
	_, err := logicalClient.Delete(identityPath)
	if err != nil {
		msg := "Error deleting policy " + policyName + " from identity " + identity + ":" + err.Error()
		return errors.New(msg)
	}

	return nil
}

// WriteVaultPolicy stores in vault the policy indicated.  This can be associated with a vault token or
// an authentication identity to ensure proper use of secrets.
// Example policy: "path \"identities/test-identity\" {\n	capabilities = [\"read\"]\n }"
// 		NOTE the line returns and the tab.  Without them it fails!
func WriteVaultPolicy(policyName string, policy string, vaultClient *api.Client) error {
	if RunWithoutVaultHook() {
		return nil
	}

	sys := vaultClient.Sys()

	err := sys.PutPolicy(policyName, policy)
	if err != nil {
		msg := "Error writing policy name " + policyName + " with rules: " + policy + ":" + err.Error()
		return errors.New(msg)
	}

	return nil
}

// DeleteVaultPolicy removes the policy with the given name from vault
func DeleteVaultPolicy(policyName string, vaultClient *api.Client) error {
	if RunWithoutVaultHook() {
		return nil
	}

	sys := vaultClient.Sys()

	err := sys.DeletePolicy(policyName)
	if err != nil {
		msg := "Error deleting policy " + policyName + ":" + err.Error()
		return errors.New(msg)
	}

	return nil
}

// DeleteFromVault deletes credentials (either user or dataset or any other)
func DeleteFromVault(vaultPath string, vaultClient *api.Client) error {
	if RunWithoutVaultHook() {
		return nil
	}

	logicalClient := vaultClient.Logical()
	if logicalClient == nil {
		msg := "No logical client received when deleting credentials from vault"
		return errors.New(msg)
	}
	_, err := logicalClient.Delete(vaultPath)
	if err != nil {
		msg := "Error deleting credentials from vault for " + vaultPath + ":" + err.Error()
		return errors.New(msg)
	}
	return nil
}

// GetFromVault returns the credentials from vault as json
func GetFromVault(vaultPath string, vaultClient *api.Client) (string, error) {
	if RunWithoutVaultHook() {
		return "", nil
	}

	logicalClient := vaultClient.Logical()
	if logicalClient == nil {
		msg := "No logical client received when retrieving credentials from vault"
		return "", errors.New(msg)
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

// GenerateUserCredentialsSecretName creates a secret name for storing user credentials in vault
func GenerateUserCredentialsSecretName(namespace string, name string, system string) string {
	return GetVaultUserHome() + namespace + "/" + name + "/" + system
}

// GetUserCredentialsFromVault is used to retrieve user credentials from vault as json
func GetUserCredentialsFromVault(namespace string, name string, system string, vaultClient *api.Client) (string, error) {
	vaultPath := GenerateUserCredentialsSecretName(namespace, name, system)
	return GetFromVault(vaultPath, vaultClient)
}

// DeleteUserCredentialsFromVault is used to delete user credentials from vault
func DeleteUserCredentialsFromVault(namespace string, name string, system string, vaultClient *api.Client) error {
	vaultPath := GenerateUserCredentialsSecretName(namespace, name, system)
	return DeleteFromVault(vaultPath, vaultClient)
}

// AddUserCredentialsToVault is used to save the credentials for accessing external systems in lieu of a user - data catalog, policy manager, credential manager
// The credentials stored are use by connectors to these systems.
func AddUserCredentialsToVault(namespace string, name string, system string, credentials map[string]interface{}, vaultClient *api.Client) (string, error) {
	vaultPath := GenerateUserCredentialsSecretName(namespace, name, system)
	if RunWithoutVaultHook() {
		return vaultPath, nil
	}

	// Add credentials to vault, and return vaultPath where they are stored
	logicalClient := vaultClient.Logical()
	if logicalClient == nil {
		msg := "No logical client received when adding user  credentials to vault"
		return vaultPath, errors.New(msg)
	}

	_, err := logicalClient.Write(vaultPath, credentials)
	if err != nil {
		msg := "Error adding credentials to vault to " + vaultPath + ":" + err.Error()
		return vaultPath, errors.New(msg)
	}
	return vaultPath, nil
}

// AddToVault adds a dataset's credentials to vault.
// The vaultClient parameter is obtained by calling InitVault
// Overwrites existing credentials with the ones provided if the credentials already exist in vault.
func AddToVault(id string, credentials map[string]interface{}, vaultClient *api.Client) (string, error) {
	vaultDatasetPath := GetVaultDatasetHome() + id
	if RunWithoutVaultHook() {
		return vaultDatasetPath, nil
	}

	// Add credentials to vault, and return vaultPath where they are stored
	logicalClient := vaultClient.Logical()
	if logicalClient == nil {
		msg := "No logical client received when adding data set credentials to vault"
		return vaultDatasetPath, errors.New(msg)
	}

	_, err := logicalClient.Write(vaultDatasetPath, credentials)
	if err != nil {
		msg := "Error adding credentials to vault to " + vaultDatasetPath + ":" + err.Error()
		return vaultDatasetPath, errors.New(msg)
	}
	return vaultDatasetPath, nil
}

// GetFullCredentialsPath returns the path to be used for credentials retrieval
func GetFullCredentialsPath(secretName string) string {
	base := GetSecretProviderURL() + "?role=" + GetSecretProviderRole() + "&secret_name="
	return fmt.Sprintf("%s%s", base, url.QueryEscape(secretName))
}

// GetUserCredentialsVaultPath returns the path that can be used externally to retrieve user credentials for a specific system and label (associated with compute)
// It is of the form <secret-provider-url>/v1/m4d-system/user-creds/{namespace}/{name}/{system}
func GetUserCredentialsVaultPath(namespace string, name string, system string) string {
	secretName := "/v1/" + GenerateUserCredentialsSecretName(namespace, name, system)
	return GetFullCredentialsPath(secretName)
}

// GetDatasetVaultPath returns the path that can be used externally to retrieve a dataset's credentials
// It is of the form <secret-provider-url>/v1/m4d-system/dataset-creds/...
func GetDatasetVaultPath(assetID string) string {
	secretName := "/v1/" + GetVaultDatasetHome() + assetID
	return GetFullCredentialsPath(secretName)
}

// GetSecretPath returns the path to the secret that holds the dataset's credentials
// It is of the form v1/m4d-system/dataset-creds/...
func GetSecretPath(assetID string) string {
	base := "v1/" + GetVaultDatasetHome()
	return fmt.Sprintf("%s%s", base, url.PathEscape(assetID))
}

// GetAuthPath returns the auth method path to use
// It is of the form v1/auth/<auth path>/login
func GetAuthPath(authPath string) string {
	fullAuthPath := fmt.Sprintf("v1/auth/%s/login", authPath)
	return fullAuthPath
}

// InitVaultAuth initiates the authentication mechanism used to obtain vault tokens.
// Vault supports multiple types of authentication.  The one used is set in the config for the m4d.
func InitVaultAuth(vaultClient *api.Client) error {
	if RunWithoutVaultHook() {
		return nil
	}

	sys := vaultClient.Sys()

	auth, options := GetVaultAuthService()
	err := sys.EnableAuthWithOptions(GetIdentitiesVaultAuthPath(), &options)
	if err != nil {
		msg := "Error enabling " + auth + " authentication for vault " + ":" + err.Error()
		return errors.New(msg)
	}

	return nil
}

// MountUserVault mounts a key-value secret provider (kv version 1) to manage the storage
// of the user credentials for the external systems accessed by the m4d
func MountUserVault(token string) error {
	if RunWithoutVaultHook() {
		return nil
	}

	body := strings.NewReader(`{"type":"kv-v1"}`)
	url := GetVaultAddress() + GetVaultUserMountPath()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		msg := "Error creating request to mount vault for user credentials " + url + ":" + err.Error()
		return errors.New(msg)
	}
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		msg := "Error mounting vault for user credentials " + url + ":" + err.Error()
		return errors.New(msg)
	}
	defer resp.Body.Close()

	return nil
}

// MountDatasetVault mounts a key-value secret provider (kv version 1) to manage the storage
// of the dataset credentials
func MountDatasetVault(token string) error {
	if RunWithoutVaultHook() {
		return nil
	}

	body := strings.NewReader(`{"type":"kv-v1"}`)
	url := GetVaultAddress() + GetVaultDatasetMountPath() // GetVaultDatasetHome()
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		msg := "Error creating request to mount vault for dataset credentials " + url + ":" + err.Error()
		return errors.New(msg)
	}
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		msg := "Error mounting vault for dataset credentials " + url + ":" + err.Error()
		return errors.New(msg)
	}
	defer resp.Body.Close()

	return nil
}

// GetIdentitiesVaultAuthPath returns the path where policies for the different identities are written
// Includes the namespace so that multiple m4d control planes can use the same vault instance.
func GetIdentitiesVaultAuthPath() string {
	auth, _ := GetVaultAuthService()
	//	path := "auth/" + auth + "/" + GetSystemNamespace() + "/identities"
	path := "auth/" + auth
	return path
}

// InitVault creates a new client for accessing and storing dataset credentials in vault.
// Note that it assumes that the dataset home path has been mounted during the vault setup.
func InitVault(token string) (*api.Client, error) {
	if RunWithoutVaultHook() {
		return nil, nil
	}
	vaultAddress := GetVaultAddress()
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

	return client, nil
}

// RunWithoutVaultHook is used for the local testing in an environment without vault service
func RunWithoutVaultHook() bool {
	return os.Getenv("RUN_WITHOUT_VAULT") == "1"
}
