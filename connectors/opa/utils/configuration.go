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
	CACERTSecretName      string = "CACERT_SECRET_NAME"
	CACERTSecretNamespace string = "CACERT_SECRET_NAMESPACE"
)

// GetCertSecretName returns the secret name which holds the
// server certficate
func GetCertSecretName() string {
	return os.Getenv(CertSecretName)
}

// GetCertSecretNamespace returns the secret name which holds the
// server certficate
func GetCertSecretNamespace() string {
	return os.Getenv(CertSecretNamespace)
}

// GetTLSEnabled returns true if the connection between
// the manager and the katalog is using tls
func GetTLSEnabled() bool {
	return os.Getenv(TLSEnabled) == "true"
}

// GetTLSEnabled returns true if the connection between
// the manager and the katalog is using mutual tls
func GetMTLSEnabled() bool {
	return os.Getenv(MTLSnabled) == "true"
}

// GetCACERTSecretName return the secret name that holds the CA certificate which is used by the server
// to validate the connection to the client if mtls is enabled.
func GetCACERTSecretName() string {
	return os.Getenv(CACERTSecretName)
}

func GetCACERTSecretNamespace() string {
	return os.Getenv(CACERTSecretNamespace)
}
