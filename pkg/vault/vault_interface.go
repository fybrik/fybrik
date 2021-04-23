// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"os"
)

// Interface provides vault functionality
type Interface interface {
	LinkPolicyToIdentity(identity string, policyName string, boundedNamespace string, serviceAccount string, auth string, ttl string) error
	RemovePolicyFromIdentity(identity string, policyName string, auth string) error
	WritePolicy(policyName string, policy string) error
	DeletePolicy(policyName string) error
	Mount(path string) error
	DeleteSecret(vaultPath string) error
	GetSecret(vaultPath string) (string, error)
	AddSecret(path string, credentials map[string]interface{}) error
	AddSecretFromStruct(path string, creds interface{}) error
}

// InitConnection creates a new connection to vault.
// Note that it assumes that the home path has been mounted during the vault setup.
func InitConnection(addr string, token string) (Interface, error) {
	if os.Getenv("RUN_WITHOUT_VAULT") == "1" {
		return NewDummyConnection()
	}
	return NewConnection(addr, token)
}
