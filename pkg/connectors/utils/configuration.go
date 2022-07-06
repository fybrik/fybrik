// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
)

// Attributes that are defined in a config map or the runtime environment
const (
	CatalogConnectorUseTLS  string = "CATALOG_CONNECTOR_USE_TLS"
	CatalogConnectorUseMTLS string = "CATALOG_CONNECTOR_USE_MTLS"
	PolicyManagerUseTLS     string = "POLICY_MANAGER_CONNECTOR_USE_TLS"
	PolicyManagerUseMTLS    string = "POLICY_MANAGER_CONNECTOR_USE_MTLS"
	CertSecretName          string = "CERT_SECRET_NAME"
	CertSecretNamespace     string = "CERT_SECRET_NAMESPACE"
	CACERTSecretName        string = "CACERT_SECRET_NAME"      //nolint:gosec
	CACERTSecretNamespace   string = "CACERT_SECRET_NAMESPACE" //nolint:gosec
)

// GetCatalogConnectorUseTLS returns true if data connector communication should use tls.
func GetCatalogConnectorUseTLS() bool {
	return os.Getenv(CatalogConnectorUseTLS) == "true"
}

// GetCatalogConnectorUseMTLS returns true if data connector communication should use mtls.
func GetCatalogConnectorUseMTLS() bool {
	return os.Getenv(CatalogConnectorUseMTLS) == "true"
}

// GetPolicyManagerUseTLS returns true if policy manager communication should use tls.
func GetPolicyManagerUseTLS() bool {
	return os.Getenv(PolicyManagerUseTLS) == "true"
}

// GetPolicyManagerUseTLS returns true if policy manager communication should use mtls.
func GetPolicyManagerUseMTLS() bool {
	return os.Getenv(PolicyManagerUseMTLS) == "true"
}

// GetCertSecretName returns the name of the kubernetes secret which holds the
// manager/connectors.
func GetCertSecretName() string {
	return os.Getenv(CertSecretName)
}

// GetCertSecretNamespace returns the namespace of the kubernetes secret which holds the
// manager/connectors.
func GetCertSecretNamespace() string {
	return os.Getenv(CertSecretNamespace)
}

// GetCACERTSecretName returns the name of the kubernetes secret that holds the CA certificates
// used by the client/server to validate the manager to the manager/connectors.
func GetCACERTSecretName() string {
	return os.Getenv(CACERTSecretName)
}

// GetCACERTSecretNamespace returns the namespace of the kubernetes secret that holds the CA certificate
// used by the client/server to validate the manager to the manager/connectors.
func GetCACERTSecretNamespace() string {
	return os.Getenv(CACERTSecretNamespace)
}
