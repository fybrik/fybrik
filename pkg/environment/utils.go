// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo"
)

// Returns the integer value of an environment variable.
// If the environment variable is not set or cannot be parsed the default value is returned.
func GetEnvAsInt(key string, defaultValue int) int {
	if env, isSet := os.LookupEnv(key); isSet {
		i, err := strconv.Atoi(env)
		if err == nil {
			return i
		}
	}
	return defaultValue
}

// Returns the float32 value of an environment variable.
// If the environment variable is not set or cannot be parsed the default value is returned.
func GetEnvAsFloat32(key string, defaultValue float32) float32 {
	if env, isSet := os.LookupEnv(key); isSet {
		f, err := strconv.ParseFloat(env, 32) //nolint:revive,gomnd // Ignore magic number 32
		if err == nil {
			return float32(f)
		}
	}
	return defaultValue
}

func MustGetEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("missing required environment variable: %s", key)
	}
	return value, nil
}

func SetIfNotSet(key, value string, t ginkgo.GinkgoTInterface) {
	if _, b := os.LookupEnv(key); !b {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Could not set environment variable %s", key)
		}
	}
}

func DefaultTestConfiguration(t ginkgo.GinkgoTInterface) {
	SetIfNotSet(CatalogConnectorServiceAddressKey, "http://localhost:50085", t)
	SetIfNotSet(VaultAddressKey, "http://127.0.0.1:8200/", t)
	SetIfNotSet(EnableWebhooksKey, "false", t)
	SetIfNotSet(ConnectionTimeoutKey, "120", t)
	SetIfNotSet(MainPolicyManagerConnectorURLKey, "http://localhost:50090", t)
	SetIfNotSet(MainPolicyManagerNameKey, "MOCK", t)
	SetIfNotSet(LoggingVerbosityKey, "-1", t)
	SetIfNotSet(PrettyLoggingKey, "true", t)
	SetIfNotSet(LocalClusterName, "thegreendragon", t)
	SetIfNotSet(LocalRegion, "theshire", t)
}

// GetSystemNamespace returns the namespace of control plane
func GetSystemNamespace() string {
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return DefaultControllerNamespace
}
