// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"os"
	"strconv"

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
	UseCSPKey                         string = "USE_CSP"
	CSPPathKey                        string = "CSP_PATH"
	DataDir                           string = "DATA_DIR"
	ModuleNamespace                   string = "MODULES_NAMESPACE"
	ControllerNamespace               string = "CONTROLLER_NAMESPACE"
	ApplicationNamespace              string = "APPLICATION_NAMESPACE"
)

// DefaultModulesNamespace defines a default namespace where module resources will be allocated
const DefaultModulesNamespace = "fybrik-blueprints"

// DefaultControllerNamespace defines a default namespace where fybrik control plane is running
const DefaultControllerNamespace = "fybrik-system"

func GetDefaultModulesNamespace() string {
	ns := os.Getenv(ModuleNamespace)
	if ns == "" {
		ns = DefaultModulesNamespace
	}
	return ns
}

func GetControllerNamespace() string {
	controllerNamespace := os.Getenv(ControllerNamespace)
	if controllerNamespace == "" {
		controllerNamespace = DefaultControllerNamespace
	}
	return controllerNamespace
}

func GetApplicationNamespace() string {
	return os.Getenv(ApplicationNamespace)
}

// GetDataDir returns the directory where the data resides.
func GetDataDir() string {
	return os.Getenv(DataDir)
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

// UseCSP return true if a CSP solver should be used when generating a plotter
func UseCSP() bool {
	return os.Getenv(UseCSPKey) == "true"
}

// GetCSPPath returns the path of the CSP solver to use when generating a plotter, or "" if no CSP solver is defined
func GetCSPPath() string {
	return os.Getenv(CSPPathKey)
}

// GetDataCatalogServiceAddress returns the address where data catalog is running
func GetDataCatalogServiceAddress() string {
	return os.Getenv(CatalogConnectorServiceAddressKey)
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
		CatalogConnectorServiceAddressKey, DataDir, ModuleNamespace, ControllerNamespace, ApplicationNamespace}

	log.Info().Msg("Manager configured with the following environment variables:")
	for _, envVar := range envVarArray {
		logEnvVariable(log, envVar)
	}
}
