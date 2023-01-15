// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"emperror.dev/errors"
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

// NewConnection returns a new Connection object
func NewConnection(addr, token string) (*Connection, error) {
	conf := &api.Config{
		Address:    addr,
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
	return &Connection{
		Client:  client,
		Address: addr,
		Token:   token,
	}, nil
}

// LinkPolicyToIdentity registers a policy for a given identity or role, meaning that when a person or service
// of that identity logs into vault and tries to read or write a secret the provided policy
// will determine whether that is allowed or not.
func (c *Connection) LinkPolicyToIdentity(identity, policyName, boundedNamespace, serviceAccount, auth, ttl string) error {
	identityPath := "auth/" + auth + "/" + identity

	logicalClient := c.Client.Logical()
	if logicalClient == nil {
		return fmt.Errorf("no logical client received when linking policy %s to idenity %s", policyName, identity)
	}

	params := map[string]interface{}{
		"user_claim":                       "sub",
		"role_type":                        auth,
		"bound_service_account_names":      serviceAccount,
		"bound_service_account_namespaces": boundedNamespace,
		"policies":                         policyName,
		"ttl":                              ttl,
	}

	_, err := logicalClient.Write(identityPath, params)
	if err != nil {
		return errors.Wrapf(err, "error linking policy %s to identity %s", policyName, identity)
	}

	return nil
}

// RemovePolicyFromIdentity removes the policy from the authentication identity with which it is associated, meaning
// this policy will no longer be invoked when a person or service authenticates with this identity.
func (c *Connection) RemovePolicyFromIdentity(identity, policyName, auth string) error {
	identityPath := "auth/" + auth + "/" + identity //nolint:revive

	logicalClient := c.Client.Logical()
	if logicalClient == nil {
		return fmt.Errorf("no logical client received when deleting policy %s from idenity %s", policyName, identity)
	}
	_, err := logicalClient.Delete(identityPath)
	if err != nil {
		return errors.Wrapf(err, "error deleting policy %s from identity %s", policyName, identity)
	}

	return nil
}

// WritePolicy stores in vault the policy indicated.  This can be associated with a vault token or
// an authentication identity to ensure proper use of secrets.
// Example policy: "path \"identities/test-identity\" {\n	capabilities = [\"read\"]\n }"
//
// NOTE the line returns and the tab.  Without them it fails!
func (c *Connection) WritePolicy(policyName, policy string) error {
	sys := c.Client.Sys()

	err := sys.PutPolicy(policyName, policy)
	if err != nil {
		return errors.Wrapf(err, "error writing policy name %s with rules: %s", policyName, policy)
	}

	return nil
}

// DeletePolicy removes the policy with the given name from vault
func (c *Connection) DeletePolicy(policyName string) error {
	sys := c.Client.Sys()

	err := sys.DeletePolicy(policyName)
	if err != nil {
		return errors.Wrapf(err, "error deleting policy %s", policyName)
	}

	return nil
}

// DeleteSecret deletes a secret
func (c *Connection) DeleteSecret(vaultPath string) error {
	logicalClient := c.Client.Logical()
	if logicalClient == nil {
		return errors.New("no logical client received when deleting credentials from vault")
	}
	_, err := logicalClient.Delete(vaultPath)
	if err != nil {
		return errors.Wrapf(err, "error deleting credentials from vault for %s", vaultPath)
	}
	return nil
}

// GetSecret returns the stored secret as json
func (c *Connection) GetSecret(vaultPath string) (string, error) {
	logicalClient := c.Client.Logical()
	if logicalClient == nil {
		return "", errors.New("no logical client received when retrieving credentials from vault")
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

// AddSecret adds a secret to vault
func (c *Connection) AddSecret(path string, credentials map[string]interface{}) error {
	logicalClient := c.Client.Logical()
	if logicalClient == nil {
		return errors.New("no logical client received when adding secrets to vault")
	}

	_, err := logicalClient.Write(path, credentials)
	if err != nil {
		return errors.Wrapf(err, "error adding credentials to vault to %s", path)
	}
	return nil
}

// AddSecretFromStruct constructs a vault secret from the given structure
func (c *Connection) AddSecretFromStruct(path string, creds interface{}) error {
	jsonStr, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	credentialsMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonStr, &credentialsMap); err != nil {
		return err
	}
	return c.AddSecret(path, credentialsMap)
}

// Mount mounts a key-value secret provider (kv version 1) to manage the storage of the secrets
func (c *Connection) Mount(path string) error {
	body := strings.NewReader(`{"type":"kv-v1"}`)
	url := c.Address + path

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return errors.Wrapf(err, "error creating request to mount vault for %s", url)
	}
	req.Header.Set("X-Vault-Token", c.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "error mounting vault for %s", url)
	}
	defer resp.Body.Close()

	return nil
}
