// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"fybrik.io/fybrik/pkg/connectors/datacatalog/clients"
	"fybrik.io/fybrik/pkg/environment"
	"github.com/gin-gonic/gin"
)

const (
	envOPAServerURL        = "OPA_SERVER_URL"
	envCatalogConnectorURL = "CATALOG_CONNECTOR_URL"
	envCatalogProviderName = "CATALOG_PROVIDER_NAME"
	envConnectionTimeout   = "CONNECTION_TIMEOUT"
)

// NewRouter returns a new router.
func NewRouter(controller *ConnectorController) *gin.Engine {
	router := gin.Default()
	router.POST("/getPoliciesDecisions", controller.GetPoliciesDecisions)
	return router
}

func main() {
	ip := ""
	port := 8080

	// Parse environment variables
	opaServerURL, err := environment.MustGetEnv(envOPAServerURL)
	if err != nil {
		log.Fatal(err.Error())
	}

	if !strings.HasPrefix(opaServerURL, "https://") && !strings.HasPrefix(opaServerURL, "http://") {
		log.Fatal("server URL for OPA server must have http or https schema")
	}

	catalogConnectorAddress, err := environment.MustGetEnv(envCatalogConnectorURL)
	if err != nil {
		log.Fatal(err.Error())
	}

	catalogProviderName, err := environment.MustGetEnv(envCatalogProviderName)
	if err != nil {
		log.Fatal(err.Error())
	}

	timeout := environment.GetEnvAsInt(envConnectionTimeout, 10)
	connectionTimeout := time.Duration(timeout) * time.Second

	// Create data catalog client
	catalogClient, err := clients.NewDataCatalog(catalogProviderName, catalogConnectorAddress, connectionTimeout)
	if err != nil {
		log.Fatal(err)
	}

	// Create and start connector
	controller := NewConnectorController(opaServerURL, catalogClient)
	router := NewRouter(controller)
	router.Use(gin.Logger())

	bindAddress := fmt.Sprintf("%s:%d", ip, port)
	err = router.Run(bindAddress)

	if err != nil {
		log.Fatal(err)
	}
}
