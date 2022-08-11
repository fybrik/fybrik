// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net/http"

	"github.com/rs/zerolog"

	fybrikTLS "fybrik.io/fybrik/pkg/tls"
)

// GetHTTPClient returns an object of type *http.Client.
func GetHTTPClient(log *zerolog.Logger) *http.Client {
	config, err := fybrikTLS.GetClientTLSConfig(log)
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
