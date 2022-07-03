// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
)

// Attributes that are defined in a config map or the runtime environment
const (
	TLSEnabled            string = "TLS_ENABLED"
	MTLSnabled            string = "MTLS_ENABLED"
	CertSecretName        string = "CERT_SECRET_NAME"
	CertSecretNamespace   string = "CERT_SECRET_NAMESPACE"
	CACERTSecretName      string = "CACERT_SECRET_NAME"      //nolint:gosec
	CACERTSecretNamespace string = "CACERT_SECRET_NAMESPACE" //nolint:gosec
)

// GetCertSecretName returns the name of the kubernetes secret which holds the
// katalog-connector certficates.
func GetCertSecretName() string {
	return os.Getenv(CertSecretName)
}

// GetCertSecretNamespace returns the namespace of the kubernetes secret which holds the
// katalog-connector certficates.
func GetCertSecretNamespace() string {
	return os.Getenv(CertSecretNamespace)
}

// GetTLSEnabled returns true if the connection between
// the manager and the katalog is using tls.
func GetTLSEnabled() bool {
	return os.Getenv(TLSEnabled) == "true"
}

// GetMTLSEnabled returns true if the connection between
// the manager and the katalog is using mutual tls authentication.
func GetMTLSEnabled() bool {
	return os.Getenv(MTLSnabled) == "true"
}

// GetCACERTSecretName returns the name of the kubernetes secret that holds the CA certificates
// used by the katalog-connector server to validate the connection to the client if mtls is enabled.
func GetCACERTSecretName() string {
	return os.Getenv(CACERTSecretName)
}

// GetCACERTSecretNamespace returns the namespace of the kubernetes secret that holds the CA certificate
// used by the katalog-connector server to validate the connection to the client if mtls is enabled.
func GetCACERTSecretNamespace() string {
	return os.Getenv(CACERTSecretNamespace)
}
