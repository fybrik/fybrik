// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"

	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/policymanager"
	fybrikTLS "fybrik.io/fybrik/pkg/tls"
)

const (
	headerCredentials = "X-Request-Cred" // #nosec G101 -- This is a false positive
	policyEndpoint    = "/v1/data/dataapi/authz/verdict"
)

type ConnectorController struct {
	OpaServerURL string
	OpaClient    *retryablehttp.Client
	Log          zerolog.Logger
}

func NewConnectorController(opaServerURL string) (*ConnectorController, error) {
	log := logging.LogInit(logging.CONNECTOR, "opa-connector")
	retryClient := retryablehttp.NewClient()
	if strings.HasPrefix(opaServerURL, "https") {
		config, err := fybrikTLS.GetClientTLSConfig(&log)
		if err != nil {
			log.Error().Err(err)
			return nil, err
		}
		if config != nil {
			log.Info().Msg("Set TLS config for opa connector as a client")
			retryClient.HTTPClient.Transport = &http.Transport{TLSClientConfig: config}
		}
	}

	return &ConnectorController{
		OpaServerURL: opaServerURL,
		OpaClient:    retryClient,
		Log:          log,
	}, nil
}

func (r *ConnectorController) GetPoliciesDecisions(c *gin.Context) {
	// Parse request
	var request policymanager.GetPolicyDecisionsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.reportError(c, http.StatusBadRequest, err.Error())
		return
	}
	logging.LogStructure("GetPoliciesDecisions object received:", request, &r.Log, zerolog.DebugLevel, false, false)
	// Add "input" hierarchy
	inputStruct := map[string]interface{}{"input": &request}
	// Marshal request as JSON
	requestBody, err := json.Marshal(&inputStruct)
	if err != nil {
		r.reportError(c, http.StatusInternalServerError, err.Error())
		return
	}
	// Send request to OPA
	endpoint := fmt.Sprintf("%s/%s", strings.TrimRight(r.OpaServerURL, "/"), strings.TrimLeft(policyEndpoint, "/"))
	responseFromOPA, err := r.OpaClient.Post(endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		r.reportError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Read response from OPA
	defer responseFromOPA.Body.Close()
	responseFromOPABody, err := io.ReadAll(responseFromOPA.Body)
	if err != nil {
		r.reportError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Handle errors from OPA
	if responseFromOPA.StatusCode != http.StatusOK {
		// TODO: better error handling for OPA errors
		r.reportError(c, responseFromOPA.StatusCode, string(responseFromOPABody))
		return
	}

	// Unmarshal as GetPolicyDecisionsResponse for the sake of validation
	var response policymanager.GetPolicyDecisionsResponse
	if err := json.Unmarshal(responseFromOPABody, &response); err != nil {
		r.reportError(c, http.StatusInternalServerError, err.Error())
		return
	}
	r.Log.Info().Msg(
		"Sending response from opa connector with created asset ID: " + string(request.Resource.ID))

	c.JSON(http.StatusOK, response)
}

func (r *ConnectorController) reportError(c *gin.Context, httpCode int, errorMessage string) {
	r.Log.Warn().CallerSkipFrame(1).Msg(errorMessage)
	c.JSON(httpCode, gin.H{"error": errorMessage})
}
