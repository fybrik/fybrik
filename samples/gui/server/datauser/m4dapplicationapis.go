// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	dm "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var dmaClient *K8sClient

// DMARoutes is a list of the REST APIs supported by the backend of the Data User GUI
func DMARoutes(client *K8sClient) *chi.Mux {
	dmaClient = client // global variable used by all funcs in this package

	router := chi.NewRouter()
	router.Get("/{m4dapplicationID}", GetM4DApplication) // Returns the M4DApplication CRD including its status
	router.Get("/", ListM4DApplications)
	router.Delete("/{m4dapplicationID}", DeleteM4DApplication)
	router.Post("/", CreateM4DApplication)
	router.Put("/{m4dapplicationID}", UpdateM4DApplication)
	router.Options("/*", M4DApplicationOptions)
	return router
}

// M4DApplicationOptions returns an OK status, but more importantly its header is set to indicate
// that future POST, PUT and DELETE calls are allowed as per the header values set when the router was initiated in main.go
func M4DApplicationOptions(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
}

// ListM4DApplications returns all of the M4DApplication instances in the namespace
func ListM4DApplications(w http.ResponseWriter, r *http.Request) {
	log.Println("In ListM4DApplications")
	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaClient set")
		}
	}

	// Call kubernetes to get the list of M4DApplication CRDs
	dmaList, err := dmaClient.ListApplications(meta_v1.ListOptions{})
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.JSON(w, r, dmaList) // Return the M4DApplication as json
}

// GetM4DApplication returns the M4DApplication CRD, both spec and status
// associated with the ID provided.
func GetM4DApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("In GetM4DApplication")
	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaclient set")
		}
	}

	m4dapplicationID := chi.URLParam(r, "m4dapplicationID")

	// Call kubernetes to get the M4DApplication CRD
	dma, err := dmaClient.GetApplication(m4dapplicationID)
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.JSON(w, r, dma) // Return the M4DApplication as json
}

// UpdateM4DApplication changes the desired state of an existing M4DApplication CRD
func UpdateM4DApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("In UpdateM4DApplication")
	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaclient set")
		}
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var dmaStruct dm.M4DApplication

	// Create the golang structure from the json
	err := decoder.Decode(&dmaStruct)
	if err != nil {
		suberr := render.Render(w, r, ErrInvalidRequest(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	m4dapplicationID := chi.URLParam(r, "m4dapplicationID")
	// Call kubernetes to update the M4DApplication CRD
	dmaStruct.Namespace = dmaClient.namespace
	dma, err := dmaClient.UpdateApplication(m4dapplicationID, &dmaStruct)
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.Status(r, http.StatusOK)
	result := DMASuccessResponse{UniqueID: dma.Name, DMA: *dma, Message: "Updated!!"}
	render.JSON(w, r, result)
}

// DeleteM4DApplication deletes the M4DApplication CRD running in the m4d control plane,
// and all of the components associated with it - ex: blueprint, modules that perform read, write, copy, transform, etc.
func DeleteM4DApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("In DeleteM4DApplication")
	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaclient set")
		}
	}

	m4dapplicationID := chi.URLParam(r, "m4dapplicationID")

	// Call kubernetes to get the M4DApplication CRD
	err := dmaClient.DeleteApplication(m4dapplicationID, nil)
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.Status(r, http.StatusOK)
	result := DMASuccessResponse{UniqueID: m4dapplicationID, Message: "Deleted!!"}
	render.JSON(w, r, result)
}

// CreateM4DApplication creates a new M4DApplication CRD with the information provided by the Data User.
// The body of the request should have a json version of the M4DApplication
// TODO - return the unique ID for the requested M4DApplication.  uniqueID=name+geography
// TODO - store the request body to file
// TODO - check if the requested M4DApplication already exists
func CreateM4DApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("In CreateM4DApplication")

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var dmaStruct dm.M4DApplication

	// Create the golang structure from the json
	err := decoder.Decode(&dmaStruct)
	if err != nil {
		suberr := render.Render(w, r, ErrInvalidRequest(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaclient set")
		}
	}

	// Create the M4DApplication CRD
	dmaStruct.Namespace = dmaClient.namespace
	dma, err := dmaClient.CreateApplication(&dmaStruct)
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.Status(r, http.StatusCreated)
	result := DMASuccessResponse{UniqueID: "123", DMA: *dma, Message: "Created!!"}
	render.JSON(w, r, result)
}

// ---------------- Responses -----------------------------------------

// DMASuccessResponse - Structure returned when REST API is successful
type DMASuccessResponse struct {
	// UniqueID of the M4DApplication
	UniqueID string `json:"uniqueID"`

	// JSON representation of the M4DApplication
	DMA dm.M4DApplication `json:"jsonDMA,omitempty"`

	// Optional message about the action performed
	Message string `json:"message,omitempty"`
}
