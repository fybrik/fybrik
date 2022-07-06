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

// GetServerTLSConfig returns the server tls config for tls connection between the manager and
// the connectors based on the following params:
// certSecretName:  kubernetes secret name which contains the server certificate
// certSecretNamespace:  kubernetes secret namespace which contains the server certificate
// caSecretName: kubernetes secret name which contains ca certificate used by the server to
// validate certificate or the client.  Used when mtls is true.
// caSecretNamespace: kubernetes secret namespace which contains ca certificate used by the server to
// validate certificate or the client. Used when mtls is true.
// mtls: true if mutual tls connection is used.
func GetServerTLSConfig(serverLog *zerolog.Logger, client kclient.Client, certSecretName, certSecretNamespace,
	caSecretName, caSecretNamespace string, mtls bool) (*tls.Config, error) {
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
	if mtls {
		serverLog.Info().Msg(MTLSEnabledMsg)
		CACertsData, err := GetCertificatesFromSecret(client, caSecretName, caSecretNamespace)
		if err != nil {
			return nil, err
		}
		CACertPool := x509.NewCertPool()
		for key, element := range CACertsData {
			// skip non cerificate keys like crt.key if exists in the secret
			if !strings.HasSuffix(key, TLSCertKeySuffix) {
				continue
			}
			if !CACertPool.AppendCertsFromPEM(element) {
				serverLog.Error().Err(err).Msg(err.Error())
				return nil, errors.New("error in GetServerTLSConfig in AppendCertsFromPEM trying to lead key:" + key)
			}
		}

		config = &tls.Config{
			Certificates: []tls.Certificate{loadedCertServer},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    CACertPool,
			MinVersion:   tls.VersionTLS13,
		}
	} else {
		serverLog.Info().Msg(MTLSDisabledMsg)
		config = &tls.Config{
			Certificates: []tls.Certificate{loadedCertServer},
			ClientAuth:   tls.NoClientCert,
			MinVersion:   tls.VersionTLS13,
		}
	}
	return config, nil
}

// GetClientTLSConfig returns the client tls config for tls connection between the manager and
// the connectors based on the following params:
// certSecretName:  kubernetes secret name which contains the client certificate
// certSecretNamespace:  kubernetes secret namespace which contains the client certificate
// caSecretName: kubernetes secret name which contains ca certificate used by the client to
// validate certificate or the server.  Used when mtls is true.
// caSecretNamespace: kubernetes secret namespace which contains ca certificate used by the client to
// validate certificate or the server. Used when mtls is true.
// mtls: true if mutual tls connection is used.
func GetClientTLSConfig(clientLog *zerolog.Logger, client kclient.Client, certSecretName, certSecretNamespace,
	caSecretName, caSecretNamespace string, mtls bool) (*tls.Config, error) {
	CACertsData, err := GetCertificatesFromSecret(client, caSecretName, caSecretNamespace)
	if err != nil {
		clientLog.Error().Err(err).Msg("error in GetCertificatesFromSecret tring to get ca cert")
		return nil, err
	}

	caCertPool := x509.NewCertPool()
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

	var tlsConfig *tls.Config
	if mtls {
		clientLog.Info().Msg(MTLSEnabledMsg)
		clientCertsData, err := GetCertificatesFromSecret(client, certSecretName, certSecretNamespace)
		if err != nil {
			clientLog.Error().Err(err).Msg("error in GetCertificatesFromSecret tring to get client/server cert")
			return nil, err
		}
		cert, err := tls.X509KeyPair(clientCertsData[corev1.TLSCertKey], clientCertsData[corev1.TLSPrivateKeyKey])
		if err != nil {
			clientLog.Error().Err(err).Msg("error in X509KeyPair")
			return nil, err
		}
		tlsConfig = &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
		}
	} else {
		clientLog.Info().Msg(MTLSDisabledMsg)
		tlsConfig = &tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS13,
		}
	}
	return tlsConfig, nil
}
