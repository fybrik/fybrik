// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"encoding/json"
	"errors"
)

// Dummy implementation for testing
type Dummy struct {
	values map[string]string
}

// NewDummyConnection returns a new Dummy object
func NewDummyConnection() *Dummy {
	return &Dummy{values: make(map[string]string)}
}

func (c *Dummy) LinkPolicyToIdentity(identity, policyName, boundedNamespace, serviceAccount, auth, ttl string) error {
	return nil
}

func (c *Dummy) RemovePolicyFromIdentity(identity, policyName, auth string) error {
	return nil
}

func (c *Dummy) WritePolicy(policyName, policy string) error {
	return nil
}

func (c *Dummy) DeletePolicy(policyName string) error {
	return nil
}

func (c *Dummy) Mount(path string) error {
	return nil
}

func (c *Dummy) DeleteSecret(vaultPath string) error {
	delete(c.values, vaultPath)
	return nil
}

func (c *Dummy) GetSecret(vaultPath string) (string, error) {
	if s, hasKey := c.values[vaultPath]; hasKey {
		return s, nil
	}
	return "", errors.New("could not find key")
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
