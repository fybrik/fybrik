// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// GetCertificatesFromSecret reads the certificates from kubernetes
// secret. Used when connection between manager and connectors uses tls.
func GetCertificatesFromSecret(log *zerolog.Logger, client kclient.Client, secretName, secretNamespace string) (map[string][]byte, error) {
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

const tlsCert = "tls.crt"
const tlsKey = "tls.key"

// GetServerTLSConfig returns the server tls config for tls connection between the manager and
// the connectors based on the following params:
// serverLog: log to use
// client: kubernetes client
// certSecretName:  name of a kubernetes secret which contains the server certificate
// certSecretNamespace:  namespace of a kubernetes secret  which contains the server certificate
// caSecretName: name of a kubernetes secret  which contains ca certificate used by the server to
// validate certificate or the client.  Used when mtls is true.
// caSecretNamespace: namespace of a kubernetes secret which contains ca certificate used by the server to
// validate certificate or the client. Used when mtls is true.
// mtls: true if mutual tls connection is used.
func GetServerTLSConfig(serverLog *zerolog.Logger, client kclient.Client, certSecretName, certSecretNamespace,
	caSecretName, caSecretNamespace string, mtls bool) (*tls.Config, error) {
	serverCertsData, err := GetCertificatesFromSecret(serverLog, client, certSecretName, certSecretNamespace)
	if err != nil {
		serverLog.Error().Msg(err.Error())
		return nil, err
	}

	loadedCertServer, err := tls.X509KeyPair(serverCertsData[tlsCert], serverCertsData[tlsKey])
	if err != nil {
		serverLog.Error().Msg(err.Error())
		return nil, err
	}
	var config *tls.Config
	if mtls {
		serverLog.Info().Msg("MTLS is enabled")
		CACertsData, err := GetCertificatesFromSecret(serverLog, client, caSecretName, caSecretNamespace)
		if err != nil {
			return nil, err
		}
		CACertPool := x509.NewCertPool()
		for _, element := range CACertsData {
			if !CACertPool.AppendCertsFromPEM(element) {
				serverLog.Error().Msg(err.Error())
				return nil, errors.New("error in GetServerTLSConfig in AppendCertsFromPEM")
			}
		}

		config = &tls.Config{
			Certificates: []tls.Certificate{loadedCertServer},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    CACertPool,
			MinVersion:   tls.VersionTLS13,
		}
	} else {
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
// clientLog: log to use
// client: kubernetes client
// certSecretName:  name of a kubernetes secret which contains the client certificate
// certSecretNamespace:  namespace of a kubernetes secret which contains the client certificate
// caSecretName: name of a kubernetes secret which contains ca certificate used by the client to
// validate certificate or the server.  Used when mtls is true.
// caSecretNamespace: namespace of a kubernetes secret which contains ca certificate used by the client to
// validate certificate or the server. Used when mtls is true.
// mtls: true if mutual tls connection is used.
func GetClientTLSConfig(clientLog *zerolog.Logger, client kclient.Client, certSecretName, certSecretNamespace,
	caSecretName, caSecretNamespace string, mtls bool) (*tls.Config, error) {
	CACertsData, err := GetCertificatesFromSecret(clientLog, client, caSecretName, caSecretNamespace)
	if err != nil {
		clientLog.Error().Err(err).Msg("error in GetCertificatesFromSecret tring to get ca cert")
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	for _, element := range CACertsData {
		if !caCertPool.AppendCertsFromPEM(element) {
			clientLog.Error().Err(err).Msg("error in AppendCertsFromPEM")
			return nil, errors.New("error in GetClientTLSConfig in AppendCertsFromPEM")
		}
	}

	var tlsConfig *tls.Config
	if mtls {
		clientLog.Info().Msg("Mutual is enabled")
		clientCertsData, err := GetCertificatesFromSecret(clientLog, client, certSecretName, certSecretNamespace)
		if err != nil {
			clientLog.Error().Err(err).Msg("error in GetCertificatesFromSecret tring to get client/server cert")
			return nil, err
		}
		cert, err := tls.X509KeyPair(clientCertsData[tlsCert], clientCertsData[tlsKey])
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
		tlsConfig = &tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS13,
		}
	}
	return tlsConfig, nil
}
