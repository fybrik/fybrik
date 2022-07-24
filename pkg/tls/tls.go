// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/pkg/environment"
)

const (
	TLSEnabledMsg   string = "TLS authentication is enabled"
	TLSDisabledMsg  string = "TLS authentication is disabled"
	MTLSEnabledMsg  string = "Mutual TLS authentication is enabled"
	MTLSDisabledMsg string = "Mutual TLS authentication is disabled"
)

// GetCertificatesFromSecret reads the certificates from kubernetes
// secret. Used when connection between manager and connectors uses tls.
func GetCertificatesFromSecret(client kclient.Client, secretName, secretNamespace string) (map[string][]byte, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secretNamespace,
			Name:      secretName,
		},
	}
	objectKey := kclient.ObjectKeyFromObject(secret)

	// Read the secret.
	err := client.Get(context.Background(), objectKey, secret)
	if err != nil {
		log.Error().Msg("Error reading cert secret data: " + err.Error())
		return nil, err
	}

	return secret.Data, nil
}

const (
	// TLSCertKeySuffix is the key suffix for tls certificates in a kubernetes secret.
	TLSCertKeySuffix = ".crt"
)

// isCertificateProvided returns true if the name and the namespace of the secret which holds
// the certificate were provided. Otherwise it returns false.
func isCertificateProvided(cerificateName, certificateNamespace string) (bool, error) {
	if cerificateName == "" && certificateNamespace == "" {
		return false, nil
	}
	if cerificateName == "" || certificateNamespace == "" {
		return false, errors.New("invalid SSL configuration found, " +
			"please set both certificate name and namespace (one is missing)")
	}
	return true, nil
}

// GetServerConfig returns the server config for tls connection between the manager and
// the connectors.
func GetServerConfig(serverLog *zerolog.Logger, client kclient.Client) (*tls.Config, error) {
	// certSecretName is kubernetes secret name which contains the server certificate
	certSecretName := environment.GetCertSecretName()
	// certSecretNamespace is kubernetes secret namespace which contains the server certificate
	certSecretNamespace := environment.GetCertSecretNamespace()
	// caSecretName is kubernetes secret name which contains ca certificate used by the server to
	// validate certificate or the client.  Used when mutual tls is used.
	caSecretName := environment.GetCACERTSecretName()
	//	caSecretNamespace is  kubernetes secret namespace which contains ca certificate used by the server to
	// validate certificate or the client. Used when mutual tls is used.
	caSecretNamespace := environment.GetCACERTSecretNamespace()
	useMTLS := environment.IsUsingMTLS()

	certProvided, err := isCertificateProvided(certSecretName, certSecretNamespace)
	if err != nil {
		serverLog.Error().Err(err).Msg("error in the provided server certificate")
		return nil, err
	}
	if !certProvided {
		// no server certs provided thus the tls is not used
		return nil, errors.New("server certificates were not provided")
	}
	serverLog.Info().Msg(TLSEnabledMsg)
	serverCertsData, err := GetCertificatesFromSecret(client, certSecretName, certSecretNamespace)
	if err != nil {
		serverLog.Error().Msg(err.Error())
		return nil, err
	}

	loadedCertServer, err := tls.X509KeyPair(serverCertsData[corev1.TLSCertKey], serverCertsData[corev1.TLSPrivateKeyKey])
	if err != nil {
		serverLog.Error().Msg(err.Error())
		return nil, err
	}
	var config *tls.Config
	if !useMTLS {
		serverLog.Info().Msg(MTLSDisabledMsg)
		config = &tls.Config{
			Certificates: []tls.Certificate{loadedCertServer},
			ClientAuth:   tls.NoClientCert,
			MinVersion:   tls.VersionTLS13,
		}
		return config, nil
	}

	serverLog.Info().Msg(MTLSEnabledMsg)
	cacertProvided, err := isCertificateProvided(caSecretName, caSecretNamespace)
	if err != nil {
		serverLog.Error().Err(err).Msg("error in the provided CA certificate")
		return nil, err
	}
	var CACertPool *x509.CertPool
	if cacertProvided {
		CACertsData, err := GetCertificatesFromSecret(client, caSecretName, caSecretNamespace)
		if err != nil {
			return nil, err
		}
		CACertPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		for key, element := range CACertsData {
			// skip non cerificate keys like crt.key if exists in the secret
			if !strings.HasSuffix(key, TLSCertKeySuffix) {
				continue
			}
			if !CACertPool.AppendCertsFromPEM(element) {
				serverLog.Error().Err(err).Msg(err.Error())
				return nil, errors.New("error in GetServerConfig in AppendCertsFromPEM trying to lead key:" + key)
			}
		}
	}

	config = &tls.Config{
		Certificates: []tls.Certificate{loadedCertServer},
		// configure mutual TLS
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  CACertPool,
		MinVersion: tls.VersionTLS13,
	}

	return config, nil
}

// GetClientTLSConfig returns the client config for tls connection between the manager and
// the connectors.
func GetClientTLSConfig(clientLog *zerolog.Logger, client kclient.Client) (*tls.Config, error) {
	// certSecretName is  kubernetes secret name which contains the client certificate
	certSecretName := environment.GetCertSecretName()
	// certSecretNamespace is kubernetes secret namespace which contains the client certificate
	certSecretNamespace := environment.GetCertSecretNamespace()
	// caSecretName is kubernetes secret name which contains ca certificate used by the client to
	// validate certificate or the server.  Used when mutual tls is used.
	caSecretName := environment.GetCACERTSecretName()
	// caSecretNamespace is kubernetes secret namespace which contains ca certificate used by the client to
	// validate certificate or the server. Used when mutual tls is used.
	caSecretNamespace := environment.GetCACERTSecretNamespace()

	var caCertPool *x509.CertPool
	cacertProvided, err := isCertificateProvided(caSecretName, caSecretNamespace)
	if err != nil {
		clientLog.Error().Err(err).Msg("error in the provided server CA certificate")
		return nil, err
	}
	if !cacertProvided {
		// no private CA certificates was provided.
		clientLog.Info().Msg("no private CA certificates were provided")
	} else {
		CACertsData, e := GetCertificatesFromSecret(client, caSecretName, caSecretNamespace)
		if e != nil {
			clientLog.Error().Err(e).Msg("error in GetCertificatesFromSecret() tring to get CA cert")
			return nil, e
		}
		clientLog.Info().Msg("private CA certificates are provided")
		caCertPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		for key, element := range CACertsData {
			// skip non cerificate keys like crt.key if exists in the secret
			if !strings.HasSuffix(key, TLSCertKeySuffix) {
				continue
			}
			if !caCertPool.AppendCertsFromPEM(element) {
				clientLog.Error().Err(err).Msg("error in AppendCertsFromPEM trying to load: " + key)
				return nil, errors.New("error in GetClientTLSConfig in AppendCertsFromPEMtrying to load: " + key)
			}
		}
	}

	var tlsConfig *tls.Config
	certProvided, err := isCertificateProvided(certSecretName, certSecretNamespace)
	if err != nil {
		clientLog.Error().Err(err).Msg("error in the provided client certificate")
		return nil, err
	}
	var cert tls.Certificate
	if certProvided {
		clientLog.Log().Msg("TLS certificates were provided for the connector")

		clientCertsData, err := GetCertificatesFromSecret(client, certSecretName, certSecretNamespace)
		if err != nil {
			clientLog.Error().Err(err).Msg("error in GetCertificatesFromSecret tring to get client/server cert")
			return nil, err
		}
		cert, err = tls.X509KeyPair(clientCertsData[corev1.TLSCertKey], clientCertsData[corev1.TLSPrivateKeyKey])
		if err != nil {
			clientLog.Error().Err(err).Msg("error in X509KeyPair")
			return nil, err
		}
	} else if !cacertProvided {
		clientLog.Log().Msg("no TLS certificates were provided for the client")
		return nil, nil
	}
	tlsConfig = &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}

	return tlsConfig, nil
}
