// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"

	"fybrik.io/fybrik/pkg/environment"
)

const (
	TLSEnabledMsg   string = "TLS authentication is enabled"
	TLSDisabledMsg  string = "TLS authentication is disabled"
	MTLSEnabledMsg  string = "Mutual TLS authentication is enabled"
	MTLSDisabledMsg string = "Mutual TLS authentication is disabled"
)

var certsDir = environment.GetDataDir() + "/tls-cert"
var cacertsDir = environment.GetDataDir() + "/tls-cacert"

var certFile = certsDir + "/" + corev1.TLSCertKey
var certPrivateKeyFile = certsDir + "/" + corev1.TLSPrivateKeyKey
var CACertFileSuffix = ".crt"

func pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// isCertificateProvided returns true if the certificate file and private key were provided as expected.
// Otherwise it returns false.
func isCertificateProvided(certFileExists, keyFileExists bool) (bool, error) {
	if certFileExists && keyFileExists {
		return true, nil
	}
	if certFileExists || keyFileExists {
		return false, errors.New("invalid SSL configuration found, " +
			"please set both certificate name and namespace (one is missing)")
	}
	return false, nil
}

// getCertificate returns a certificate for the server/client if such provided.
func getCertificate() (*tls.Certificate, error) {
	// Mounted cert files.
	certFileExists := pathExists(certFile)
	keyFileExists := pathExists(certPrivateKeyFile)

	certProvided, err := isCertificateProvided(certFileExists, keyFileExists)
	if err != nil {
		return nil, err
	}
	if !certProvided {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, certPrivateKeyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func find(root, ext string) ([]string, error) {
	var a []string
	dirExists := pathExists(cacertsDir)
	if !dirExists {
		return nil, nil
	}
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// getCACertPool returns the CA certificate if such provided.
func getCACertPool() (*x509.CertPool, error) {
	var CACertPool *x509.CertPool
	var err error
	certFiles, err := find(cacertsDir, CACertFileSuffix)
	if err != nil {
		return nil, err
	}
	if certFiles != nil {
		CACertPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		for _, cacertFile := range certFiles {
			caCert, err := os.ReadFile(cacertFile)
			if err != nil {
				return nil, err
			}
			if !CACertPool.AppendCertsFromPEM(caCert) {
				return nil, errors.New("error in AppendCertsFromPEM trying to load CA certificate")
			}
		}
		return CACertPool, nil
	}
	return nil, nil
}

// GetServerConfig returns the server config for tls connection between the manager and
// the connectors.
func GetServerConfig(serverLog *zerolog.Logger) (*tls.Config, error) {
	useMTLS := environment.IsUsingMTLS()
	var err error

	loadedCertServer, err := getCertificate()
	if err != nil {
		serverLog.Error().Msg(err.Error())
		return nil, err
	}
	if loadedCertServer == nil {
		return nil, errors.New("server certificate is missing")
	}
	serverLog.Info().Msg(TLSEnabledMsg)
	var config *tls.Config
	if !useMTLS {
		serverLog.Info().Msg(MTLSDisabledMsg)
		//nolint:gosec // ignore G402: TLS MinVersion too low
		config = &tls.Config{
			Certificates: []tls.Certificate{*loadedCertServer},
			// Do not use mutual TLS
			ClientAuth: tls.NoClientCert,
			MinVersion: environment.GetTLSMinVersion(serverLog),
		}
		return config, nil
	}
	serverLog.Info().Msg(MTLSEnabledMsg)
	caCertPool, err := getCACertPool()
	if err != nil {
		return nil, err
	}
	if caCertPool != nil {
		serverLog.Log().Msg("private CA certificates were provided in GetServerConfig")
	}
	//nolint:gosec // ignore G402: TLS MinVersion too low
	config = &tls.Config{
		Certificates: []tls.Certificate{*loadedCertServer},
		// configure mutual TLS
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  caCertPool,
		MinVersion: environment.GetTLSMinVersion(serverLog),
	}

	return config, nil
}

// GetClientTLSConfig returns the client config for tls connection between the manager and
// the connectors.
func GetClientTLSConfig(clientLog *zerolog.Logger) (*tls.Config, error) {
	caCertPool, err := getCACertPool()
	if err != nil {
		return nil, err
	} else if caCertPool != nil {
		clientLog.Log().Msg("private CA certificates were provided in GetClientTLSConfig")
	} else {
		clientLog.Log().Msg("private CA certificates were not provided in GetClientTLSConfig")
	}
	cert, err := getCertificate()
	if err != nil {
		clientLog.Error().Msg(err.Error())
		return nil, err
	}
	if cert == nil && caCertPool == nil {
		clientLog.Log().Msg("no TLS certificates were provided")
		return nil, nil
	}
	var providedCert tls.Certificate
	if cert != nil {
		clientLog.Log().Msg("client TLS certificates were provided")
		providedCert = *cert
	}

	//nolint:gosec // ignore G402: TLS MinVersion too low
	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{providedCert},
		MinVersion:   environment.GetTLSMinVersion(clientLog),
	}

	return tlsConfig, nil
}
