// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"

	fybrikTLS "fybrik.io/fybrik/pkg/tls"
)

// GetHTTPClient returns an object of type *retryablehttp.Client.
func GetHTTPClient(log *zerolog.Logger) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	config, err := fybrikTLS.GetClientTLSConfig(log)
	if err != nil {
		log.Error().Err(err)
		return nil
	}
	if config != nil {
		retryClient.HTTPClient.Transport = &http.Transport{TLSClientConfig: config}
	}
	return retryClient
}
