// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import "encoding/json"

// Dummy implementation for testing
type Dummy struct {
	values map[string]string
}

// NewDummyConnection returns a new Dummy object
func NewDummyConnection() (*Dummy, error) {
	return &Dummy{values: make(map[string]string)}, nil
}

func (c *Dummy) LinkPolicyToIdentity(identity string, policyName string, boundedNamespace string, serviceAccount string, auth string, ttl string) error {
	return nil
}

func (c *Dummy) RemovePolicyFromIdentity(identity string, policyName string, auth string) error {
	return nil
}

func (c *Dummy) WritePolicy(policyName string, policy string) error {
	return nil
}

func (c *Dummy) DeletePolicy(policyName string) error {
	return nil
}

func (c *Dummy) Mount(path string) error {
	return nil
}

func (c *Dummy) DeleteSecret(vaultPath string) error {
	c.values[vaultPath] = ""
	return nil
}

func (c *Dummy) GetSecret(vaultPath string) (string, error) {
	return c.values[vaultPath], nil
}

func (c *Dummy) AddSecret(path string, credentials map[string]interface{}) error {
	bytes, err := json.Marshal(credentials)
	if err != nil {
		return err
	}
	c.values[path] = string(bytes)
	return nil
}

func (c *Dummy) AddSecretFromStruct(path string, creds interface{}) error {
	bytes, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	c.values[path] = string(bytes)
	return nil
}
