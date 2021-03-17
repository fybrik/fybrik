// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	datauser "github.com/ibm/the-mesh-for-data/samples/gui/server/datauser"
)

// Routes are the REST endpoints for CRUD operations on M4DApplication CRDS
func Routes(k8sclient *datauser.K8sClient) *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		middleware.SetHeader("Access-Control-Allow-Origin", "*"), // Allow any client to access these APIs
		middleware.SetHeader("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE"),
		middleware.SetHeader("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization"),
		render.SetContentType(render.ContentTypeJSON), // Set content-Type headers as application/json
		middleware.Logger, // Log API request calls
		//		middleware.Compress,        // Compress results, mostly gzipping assets and json
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
	)

	router.Route("/v1/dma", func(r chi.Router) {
		r.Mount("/m4dapplication", datauser.DMARoutes(k8sclient))
	})

	router.Route("/v1/creds", func(r chi.Router) {
		r.Mount("/usercredentials", datauser.CredentialRoutes(k8sclient))
	})

	router.Route("/v1/env", func(r chi.Router) {
		r.Mount("/datauserenv", datauser.EnvironmentRoutes())
	})

	return router
}

// Assumes that the control plane exists, i.e. M4DApplication custom resource exists in the cluster
func main() {
	// Init kubernetes config.  It is assumed that the GUI runs inside the same cluster and namespace
	// as the compute deployed by the user.
	k8sclient, err := datauser.K8sInit()
	if err != nil {
		panic(err.Error())
	}

	if k8sclient == nil {
		panic("Failed getting kubernetes client!")
	}

	// REST APIs provided
	router := Routes(k8sclient)
	//	credrouter := CredentialRoutes()

	// Print out all the APIs
	log.Printf("Server listening on port 8080")
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) // Walk and print out all routes
		return nil
	}

	// Print out M4DApplication APIs
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	// Print out credential APIs
	//	if err := chi.Walk(credrouter, walkFunc); err != nil {
	//		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	//	}

	log.Fatal(http.ListenAndServe(":8080", router)) // Note, the port is usually gotten from the environment.
}
