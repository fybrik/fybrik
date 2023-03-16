// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Attributes that are defined in a config map or the runtime environment
const (
	CatalogConnectorServiceAddressKey string = "CATALOG_CONNECTOR_URL"
	StorageManagerAddressKey          string = "STORAGE_MANAGER_URL"
	VaultEnabledKey                   string = "VAULT_ENABLED"
	VaultAddressKey                   string = "VAULT_ADDRESS"
	VaultModulesRoleKey               string = "VAULT_MODULES_ROLE"
	EnableWebhooksKey                 string = "ENABLE_WEBHOOKS"
	MainPolicyManagerNameKey          string = "MAIN_POLICY_MANAGER_NAME"
	MainPolicyManagerConnectorURLKey  string = "MAIN_POLICY_MANAGER_CONNECTOR_URL"
	LoggingVerbosityKey               string = "LOGGING_VERBOSITY"
	PrettyLoggingKey                  string = "PRETTY_LOGGING"
	CatalogProviderNameKey            string = "CATALOG_PROVIDER_NAME"
	DatapathLimitKey                  string = "DATAPATH_LIMIT"
	UseCSPKey                         string = "USE_CSP"
	CSPPathKey                        string = "CSP_PATH"
	CSPArgsKey                        string = "CSP_ARGS"
	DataDir                           string = "DATA_DIR"
	ModuleNamespace                   string = "MODULES_NAMESPACE"
	ControllerNamespace               string = "CONTROLLER_NAMESPACE"
	ApplicationNamespace              string = "APPLICATION_NAMESPACE"
	AdminCRsNamespace                 string = "ADMIN_CRS_NAMESPACE"
	InternalCRsNamespace              string = "INTERNAL_CRS_NAMESPACE"
	UseTLS                            string = "USE_TLS"
	UseMTLS                           string = "USE_MTLS"
	MinTLSVersion                     string = "MIN_TLS_VERSION"
	LocalClusterName                  string = "ClusterName"
	LocalZone                         string = "Zone"
	LocalRegion                       string = "Region"
	LocalVaultAuthPath                string = "VaultAuthPath"
	ResourcesPollingInterval          string = "RESOURCE_POLLING_INTERVAL"
	DiscoveryBurst                    string = "DISCOVERY_BURST"
	DiscoveryQPS                      string = "DISCOVERY_QPS"
	NPEnabled                         string = "NP_Enabled"
)

const printValueStr = "%s set to \"%s\""

// DefaultModulesNamespace defines a default namespace where module resources will be allocated
const DefaultModulesNamespace = "fybrik-blueprints"

// DefaultControllerNamespace defines a default namespace where fybrik control plane is running
const DefaultControllerNamespace = "fybrik-system"

// defaultPollingInterval defines the default time interval to check the status of the resources
// deployed by the manager. The interval is specified in milliseconds.
const defaultPollingInterval = 2000 * time.Millisecond

func GetLocalClusterName() string {
	return os.Getenv(LocalClusterName)
}

func GetLocalZone() string {
	return os.Getenv(LocalZone)
}

func GetLocalRegion() string {
	return os.Getenv(LocalRegion)
}

func GetLocalVaultAuthPath() string {
	return os.Getenv(LocalVaultAuthPath)
}

func GetCatalogProvider() string {
	return os.Getenv(CatalogProviderNameKey)
}

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

func GetInternalCRsNamespace() string {
	internalCRsNamespace := os.Getenv(InternalCRsNamespace)
	if internalCRsNamespace == "" {
		internalCRsNamespace = GetControllerNamespace()
	}
	return internalCRsNamespace
}

func GetAdminCRsNamespace() string {
	adminCRsNamespace := os.Getenv(AdminCRsNamespace)
	if adminCRsNamespace == "" {
		adminCRsNamespace = GetControllerNamespace()
	}
	return adminCRsNamespace
}

// IsUsingTLS returns true if the connector communication should use tls.
func IsUsingTLS() bool {
	return strings.ToLower(os.Getenv(UseTLS)) == "true"
}

// IsUsingMTLS returns true if the connector communication should use mtls.
func IsUsingMTLS() bool {
	return strings.ToLower(os.Getenv(UseMTLS)) == "true"
}

// GetMinTLSVersion returns the minimum TLS version that is acceptable.
// if not provided it returns zero which means that
// the system default value is used.
func GetMinTLSVersion(log *zerolog.Logger) uint16 {
	minVersion := os.Getenv(MinTLSVersion)
	rv := uint16(0)
	switch minVersion {
	case "TLS-1.0":
		rv = tls.VersionTLS10
	case "TLS-1.1":
		rv = tls.VersionTLS11
	case "TLS-1.2":
		rv = tls.VersionTLS12
	case "TLS-1.3":
		rv = tls.VersionTLS13
	default:
		log.Info().Msg("MinTLSVersion is set to the system default value")
		return rv
	}
	log.Info().Msg("MinTLSVersion is set to " + minVersion)
	return rv
}

func isNPEnabled() bool {
	return strings.ToLower(os.Getenv(NPEnabled)) == "true"
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

// GetResourcesPollingInterval returns the time interval to check the
// status of the resources deployed by the manager. The interval is specified
// in milliseconds.
// The function returns a default value if an error occurs or
// if ResourcesPollingInterval env var is undefined.
func GetResourcesPollingInterval() (time.Duration, error) {
	intervalStr := os.Getenv(ResourcesPollingInterval)
	if intervalStr == "" {
		return defaultPollingInterval, nil
	}
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return defaultPollingInterval, err
	}
	return time.Duration(interval) * time.Millisecond, nil
}

// GetDiscoveryBurst returns the K8s discovery burst value if it is set, otherwise it returns -1
func GetDiscoveryBurst() (int, error) {
	burstStr := os.Getenv(DiscoveryBurst)
	if burstStr == "" {
		return -1, nil
	}
	burst, err := strconv.Atoi(burstStr)
	if err != nil {
		return -1, err
	}
	if burst <= 0 {
		return -1, fmt.Errorf("discovery burst should be positive, got %d", burst)
	}
	return burst, err
}

// GetDiscoveryQPS returns the K8s discovery QPS value if it is set, otherwise it returns -1
func GetDiscoveryQPS() (float32, error) {
	qpsStr := os.Getenv(DiscoveryQPS)
	if qpsStr == "" {
		return -1, nil
	}
	//nolint:revive,gomnd // ignore magic numbers
	qps, err := strconv.ParseFloat(qpsStr, 32)
	if err != nil {
		return -1, err
	}
	if qps <= 0 {
		return -1, fmt.Errorf("discovery QPS should be positive, got %f", qps)
	}
	return float32(qps), err
}

// GetVaultAddress returns the address and port of the vault system,
// which is used for managing data set credentials
func GetVaultAddress() string {
	return os.Getenv(VaultAddressKey)
}

// GetDataPathMaxSize bounds the data path size (number of modules that access data for read/write/copy,
// not including transformations)
// The function returns a default value if an error occurs or if DatapathLimitKey env var
// is undefined.
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

// GetCSPArgs returns CSP solver arguments
func GetCSPArgs() string {
	return os.Getenv(CSPArgsKey)
}

// GetDataCatalogServiceAddress returns the address where data catalog is running
func GetDataCatalogServiceAddress() string {
	return os.Getenv(CatalogConnectorServiceAddressKey)
}

// GetStorageManagerAddress returns the address of storage manager
func GetStorageManagerAddress() string {
	return os.Getenv(StorageManagerAddressKey)
}

func logEnvVariable(log *zerolog.Logger, key string) {
	value, found := os.LookupEnv(key)
	if found {
		log.Info().Msgf(printValueStr, key, value)
	} else {
		log.Info().Msgf("%s is undefined", key)
	}
}

// logEnvVarUpdatedValue logs environment variables values that might be different
// from the values they were originally set to.
func logEnvVarUpdatedValue(log *zerolog.Logger, envVar, value string, err error) {
	if err != nil {
		log.Warn().Msg("error getting " + envVar + ". Setting the default to " +
			value)
		return
	}
	log.Info().Msgf(printValueStr, envVar, value)
}

func LogEnvVariables(log *zerolog.Logger) {
	envVarArray := [...]string{CatalogConnectorServiceAddressKey, StorageManagerAddressKey, VaultAddressKey, VaultModulesRoleKey,
		EnableWebhooksKey, MainPolicyManagerConnectorURLKey,
		MainPolicyManagerNameKey, LoggingVerbosityKey, PrettyLoggingKey,
		DataDir, ModuleNamespace, ControllerNamespace, ApplicationNamespace, MinTLSVersion, NPEnabled}

	log.Info().Msg("Manager configured with the following environment variables:")
	for _, envVar := range envVarArray {
		logEnvVariable(log, envVar)
	}

	interval, err := GetResourcesPollingInterval()
	logEnvVarUpdatedValue(log, ResourcesPollingInterval, interval.String(), err)
	discoveryBurst, err := GetDiscoveryBurst()
	logEnvVarUpdatedValue(log, DiscoveryBurst, strconv.Itoa(discoveryBurst), err)
	discoveryQPS, err := GetDiscoveryQPS()
	logEnvVarUpdatedValue(log, DiscoveryQPS, fmt.Sprintf("%f", discoveryQPS), err)
	dataPathMaxSize, err := GetDataPathMaxSize()
	logEnvVarUpdatedValue(log, DatapathLimitKey, strconv.Itoa(dataPathMaxSize), err)
}
