// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// EnvironmentInfo contains the info about the user's cluster/namespace and the external systems used by Fybrik
type EnvironmentInfo struct {
	// Namespace in which the GUI and GUI server are running
	Namespace string `json:"namespace"`

	// Geography in which the GUI and GUI server are running
	Geography string `json:"geography"`

	// Systems and the credentials they require
	Systems map[string][]string `json:"systems"`

	// DataSetIDStruct format which must be provided to Fybrik
	DataSetIDFormat string `json:"dataSetIDFormat"`
}

// EnvironmentRoutes provide information about the cluster/namespace in which the GUI is running
// as well as info about the Fybrik control plane deployment assumptions - ex: Data Catalog in use
func EnvironmentRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", GetEnvInfo)
	router.Options("/", EnvOptions)
	return router
}

// EnvOptions returns an OK status, but more importantly its header is set to indicate
// that future POST, PUT and DELETE calls are allowed as per the header values set when the router was initiated in main.go
func EnvOptions(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
}

// GetEnvInfo provide information about the cluster/namespace in which the GUI is running
// as well as info about the Fybrik control plane deployment assumptions - ex: Data Catalog in use
func GetEnvInfo(w http.ResponseWriter, r *http.Request) {
	var envInfo EnvironmentInfo
	// Get the geography info from the environment variable
	envInfo.Geography = os.Getenv("GEOGRAPHY")

	// Get the namespace in which we are running
	envInfo.Namespace = GetCurrentNamespace()
	render.JSON(w, r, envInfo) // Return the environment info as json
}
