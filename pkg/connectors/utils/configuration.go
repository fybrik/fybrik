// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
)

// Attributes that are defined in a config map or the runtime environment
const (
	isTLSPort             string = "CONNECTOR_USE_TLS"
	ConnectorUseMTLS      string = "CONNECTOR_USE_MTLS"
	CertSecretName        string = "CERT_SECRET_NAME"
	CertSecretNamespace   string = "CERT_SECRET_NAMESPACE"
	CACERTSecretName      string = "CACERT_SECRET_NAME"      //nolint:gosec
	CACERTSecretNamespace string = "CACERT_SECRET_NAMESPACE" //nolint:gosec
)

// isTLSPort returns true if the connector communication should use tls.
func GetisTLSPort() bool {
	return os.Getenv(isTLSPort) == "true"
}

// GetConnectorUseMTLS returns true if the connector communication should use mtls.
func GetConnectorUseMTLS() bool {
	return os.Getenv(ConnectorUseMTLS) == "true"
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
