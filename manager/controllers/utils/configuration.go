// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/rs/zerolog"
)

// Attributes that are defined in a config map or the runtime environment
const (
	CatalogConnectorServiceAddressKey string = "CATALOG_CONNECTOR_URL"
	VaultEnabledKey                   string = "VAULT_ENABLED"
	VaultAddressKey                   string = "VAULT_ADDRESS"
	VaultModulesRoleKey               string = "VAULT_MODULES_ROLE"
	EnableWebhooksKey                 string = "ENABLE_WEBHOOKS"
	ConnectionTimeoutKey              string = "CONNECTION_TIMEOUT"
	MainPolicyManagerNameKey          string = "MAIN_POLICY_MANAGER_NAME"
	MainPolicyManagerConnectorURLKey  string = "MAIN_POLICY_MANAGER_CONNECTOR_URL"
	LoggingVerbosityKey               string = "LOGGING_VERBOSITY"
	PrettyLoggingKey                  string = "PRETTY_LOGGING"
	CatalogProviderNameKey            string = "CATALOG_PROVIDER_NAME"
	DatapathLimitKey                  string = "DATAPATH_LIMIT"
	CSPPathKey                        string = "CSP_PATH"
)

// GetSystemNamespace returns the namespace of control plane
func GetSystemNamespace() string {
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return DefaultControllerNamespace
}

func IsVaultEnabled() bool {
	v := os.Getenv(VaultEnabledKey)
	return v == "true"
}

// GetModulesRole returns the modules assigned authentication role for accessing dataset credentials
func GetModulesRole() string {
	return os.Getenv(VaultModulesRoleKey)
}

// GetVaultAddress returns the address and port of the vault system,
// which is used for managing data set credentials
func GetVaultAddress() string {
	return os.Getenv(VaultAddressKey)
}

// GetDataPathMaxSize bounds the data path size (number of modules that access data for read/write/copy, not including transformations)
func GetDataPathMaxSize() (int, error) {
	defaultLimit := 2
	limitStr := os.Getenv(DatapathLimitKey)
	if limitStr == "" {
		return defaultLimit, nil
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return defaultLimit, err
	}
	return limit, nil
}

// GetCSPPath returns the path of the CSP solver to use when generating a plotter, or "" if no CSP solver should be used
func GetCSPPath() string {
	return os.Getenv(CSPPathKey)
}

// GetDataCatalogServiceAddress returns the address where data catalog is running
func GetDataCatalogServiceAddress() string {
	return os.Getenv(CatalogConnectorServiceAddressKey)
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
}

func logEnvVariable(log *zerolog.Logger, key string) {
	value, found := os.LookupEnv(key)
	if found {
		log.Info().Msgf("%s set to \"%s\"", key, value)
	} else {
		log.Info().Msgf("%s is undefined", key)
	}
}

func LogEnvVariables(log *zerolog.Logger) {
	envVarArray := [...]string{CatalogConnectorServiceAddressKey, VaultAddressKey, VaultModulesRoleKey,
		EnableWebhooksKey, ConnectionTimeoutKey, MainPolicyManagerConnectorURLKey,
		MainPolicyManagerNameKey, LoggingVerbosityKey, PrettyLoggingKey, DatapathLimitKey,
		CatalogConnectorServiceAddressKey}

	log.Info().Msg("Manager configured with the following environment variables:")
	for _, envVar := range envVarArray {
		logEnvVariable(log, envVar)
	}
}
