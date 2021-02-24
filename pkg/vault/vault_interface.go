// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Connection contains required information for connecting to vault
type Connection struct {
	Client  *api.Client
	Address string
	Token   string
}

// Dummy implementation for testing
type Dummy struct {
}

// VaultInterface provides vault functionality
type VaultInterface interface {
	LinkVaultPolicyToIdentity(identity string, policyName string, boundedNamespace string) error
	RemoveVaultPolicyFromIdentity(identity string, policyName string) error
	WriteVaultPolicy(policyName string, policy string) error
	DeleteVaultPolicy(policyName string) error
	Mount(path string) error
	DeleteSecret(vaultPath string) error
	GetSecret(vaultPath string) (string, error)
	AddSecret(path string, credentials map[string]interface{}) error
}

// LinkVaultPolicyToIdentity registers a policy for a given identity or role, meaning that when a person or service
// of that identity logs into vault and tries to read or write a secret the provided policy
// will determine whether that is allowed or not.
func (c *Connection) LinkVaultPolicyToIdentity(identity string, policyName string, boundedNamespace string) error {
	auth, _ := GetVaultAuthService()
	identityPath := GetIdentitiesVaultAuthPath() + "/" + identity

	logicalClient := c.Client.Logical()
	if logicalClient == nil {
		msg := "No logical client received when linking policy " + policyName + " to idenity " + identity
		return errors.New(msg)
	}

	params := map[string]interface{}{
		"user_claim":                       "sub",
		"role_type":                        auth,
		"bound_service_account_names":      "secret-provider",
		"bound_service_account_namespaces": boundedNamespace,
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
func (c *Connection) RemoveVaultPolicyFromIdentity(identity string, policyName string) error {
	identityPath := GetIdentitiesVaultAuthPath() + "/" + identity

	logicalClient := c.Client.Logical()
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
func (c *Connection) WriteVaultPolicy(policyName string, policy string) error {
	sys := c.Client.Sys()

	err := sys.PutPolicy(policyName, policy)
	if err != nil {
		msg := "Error writing policy name " + policyName + " with rules: " + policy + ":" + err.Error()
		return errors.New(msg)
	}

	return nil
}

// DeleteVaultPolicy removes the policy with the given name from vault
func (c *Connection) DeleteVaultPolicy(policyName string) error {
	sys := c.Client.Sys()

	err := sys.DeletePolicy(policyName)
	if err != nil {
		msg := "Error deleting policy " + policyName + ":" + err.Error()
		return errors.New(msg)
	}

	return nil
}

// DeleteSecret deletes a secret
func (c *Connection) DeleteSecret(vaultPath string) error {
	logicalClient := c.Client.Logical()
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

// GetSecret returns the stored secret as json
func (c *Connection) GetSecret(vaultPath string) (string, error) {
	logicalClient := c.Client.Logical()
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

// AddSecret adds a secret to vault
func (c *Connection) AddSecret(path string, credentials map[string]interface{}) error {
	logicalClient := c.Client.Logical()
	if logicalClient == nil {
		msg := "No logical client received when adding secrets to vault"
		return errors.New(msg)
	}

	_, err := logicalClient.Write(path, credentials)
	if err != nil {
		msg := "Error adding credentials to vault to " + path + ":" + err.Error()
		return errors.New(msg)
	}
	return nil
}

// Mount mounts a key-value secret provider (kv version 1) to manage the storage of the secrets
func (c *Connection) Mount(path string) error {
	body := strings.NewReader(`{"type":"kv-v1"}`)
	url := c.Address + path

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		msg := "Error creating request to mount vault for user credentials " + url + ":" + err.Error()
		return errors.New(msg)
	}
	req.Header.Set("X-Vault-Token", c.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		msg := "Error mounting vault for user credentials " + url + ":" + err.Error()
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

// InitConnection creates a new connection to vault.
// Note that it assumes that the home path has been mounted during the vault setup.
func InitConnection(addr string, token string) (VaultInterface, error) {
	if os.Getenv("RUN_WITHOUT_VAULT") == "1" {
		return &Dummy{}, nil
	}
	conf := &api.Config{
		Address:    addr,
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

	return &Connection{Address: addr, Token: token, Client: client}, nil
}

// GetVaultAuthTTL returns the amount of time the authorization issued by vault is valid
func GetVaultAuthTTL() string {
	return os.Getenv("VAULT_TTL")
}

// GetVaultAuthService returns the authentication service that was chosen for use with vault,
// and the configuration options for it.
// Vault support multiple different types of authentication such as java web tokens (jwt), github, aws ...
func GetVaultAuthService() (string, api.EnableAuthOptions) {
	auth := os.Getenv("VAULT_AUTH")
	options := api.EnableAuthOptions{
		Type: auth,
	}
	return auth, options
}

func (c *Dummy) LinkVaultPolicyToIdentity(identity string, policyName string, boundedNamespace string) error {
	return nil
}

func (c *Dummy) RemoveVaultPolicyFromIdentity(identity string, policyName string) error {
	return nil
}

func (c *Dummy) WriteVaultPolicy(policyName string, policy string) error {
	return nil
}

func (c *Dummy) DeleteVaultPolicy(policyName string) error {
	return nil
}

func (c *Dummy) Mount(path string) error {
	return nil
}

func (c *Dummy) DeleteSecret(vaultPath string) error {
	return nil
}

func (c *Dummy) GetSecret(vaultPath string) (string, error) {
	return "", nil
}

func (c *Dummy) AddSecret(path string, credentials map[string]interface{}) error {
	return nil
}

// Helper functions

// AddSecretFromStruct constructs a vault secret from the given structure
func AddSecretFromStruct(path string, creds interface{}, conn VaultInterface) error {
	jsonStr, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	credentialsMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonStr, &credentialsMap); err != nil {
		return err
	}
	return conn.AddSecret(path, credentialsMap)
}

