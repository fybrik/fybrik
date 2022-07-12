// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net/http"

	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	fybrikTLS "fybrik.io/fybrik/pkg/tls"
)

// GetHTTPClient returns an object of type *http.Client.
func GetHTTPClient(log *zerolog.Logger) *http.Client {
	scheme := runtime.NewScheme()
	log.Info().Msg(fybrikTLS.TLSEnabledMsg)
	err := corev1.AddToScheme(scheme)
	if err != nil {
		log.Error().Err(err)
		return nil
	}

	client, err := kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: scheme})
	if err != nil {
		log.Error().Err(err)
		return nil
	}
	config, err := fybrikTLS.GetClientTLSConfig(log, client)
	if err != nil {
		log.Error().Err(err)
		return nil
	}
	if config != nil {
		transport := &http.Transport{TLSClientConfig: config}
		return &http.Client{Transport: transport}
	}
	return http.DefaultClient
}
