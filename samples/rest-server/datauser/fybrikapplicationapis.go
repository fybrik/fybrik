// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	dm "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var dmaClient *K8sClient

// DMARoutes is a list of the REST APIs supported by the backend of the Data User GUI
func DMARoutes(client *K8sClient) *chi.Mux {
	dmaClient = client // global variable used by all funcs in this package

	router := chi.NewRouter()
	router.Get("/{fybrikapplicationID}", GetFybrikApplication) // Returns the FybrikApplication CRD including its status
	router.Get("/", ListFybrikApplications)
	router.Delete("/{fybrikapplicationID}", DeleteFybrikApplication)
	router.Post("/", CreateFybrikApplication)
	router.Put("/{fybrikapplicationID}", UpdateFybrikApplication)
	router.Options("/*", FybrikApplicationOptions)
	return router
}

// FybrikApplicationOptions returns an OK status, but more importantly its header is set to indicate
// that future POST, PUT and DELETE calls are allowed as per the header values set when the router was initiated in main.go
func FybrikApplicationOptions(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
}

// ListFybrikApplications returns all of the FybrikApplication instances in the namespace
func ListFybrikApplications(w http.ResponseWriter, r *http.Request) {
	log.Println("In ListFybrikApplications")
	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("no dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaClient set")
		}
	}

	// Call kubernetes to get the list of FybrikApplication CRDs
	dmaList, err := dmaClient.ListApplications(meta_v1.ListOptions{})
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.JSON(w, r, dmaList) // Return the FybrikApplication as json
}

// GetFybrikApplication returns the FybrikApplication CRD, both spec and status
// associated with the ID provided.
func GetFybrikApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("In GetFybrikApplication")
	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("no dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaclient set")
		}
	}

	fybrikapplicationID := chi.URLParam(r, "fybrikapplicationID")

	// Call kubernetes to get the FybrikApplication CRD
	dma, err := dmaClient.GetApplication(fybrikapplicationID)
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.JSON(w, r, dma) // Return the FybrikApplication as json
}

// UpdateFybrikApplication changes the desired state of an existing FybrikApplication CRD
func UpdateFybrikApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("In UpdateFybrikApplication")
	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("no dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaclient set")
		}
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var dmaStruct dm.FybrikApplication

	// Create the golang structure from the json
	err := decoder.Decode(&dmaStruct)
	if err != nil {
		suberr := render.Render(w, r, ErrInvalidRequest(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	fybrikapplicationID := chi.URLParam(r, "fybrikapplicationID")
	// Call kubernetes to update the FybrikApplication CRD
	dmaStruct.Namespace = dmaClient.namespace
	dma, err := dmaClient.UpdateApplication(fybrikapplicationID, &dmaStruct)
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

// DeleteFybrikApplication deletes the FybrikApplication CRD running in the fybrik control plane,
// and all of the components associated with it - ex: blueprint, modules that perform read, write, copy, transform, etc.
func DeleteFybrikApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("In DeleteFybrikApplication")
	if dmaClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("no dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaclient set")
		}
	}

	fybrikapplicationID := chi.URLParam(r, "fybrikapplicationID")

	// Call kubernetes to get the FybrikApplication CRD
	err := dmaClient.DeleteApplication(fybrikapplicationID, nil)
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.Status(r, http.StatusOK)
	result := DMASuccessResponse{UniqueID: fybrikapplicationID, Message: "Deleted!!"}
	render.JSON(w, r, result)
}

// CreateFybrikApplication creates a new FybrikApplication CRD with the information provided by the Data User.
// The body of the request should have a json version of the FybrikApplication
// TODO - return the unique ID for the requested FybrikApplication.  uniqueID=name+geography
// TODO - store the request body to file
// TODO - check if the requested FybrikApplication already exists
func CreateFybrikApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("In CreateFybrikApplication")

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var dmaStruct dm.FybrikApplication

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
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("no dmaClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no dmaclient set")
		}
	}

	// Create the FybrikApplication CRD
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
	// UniqueID of the FybrikApplication
	UniqueID string `json:"uniqueID"`

	// JSON representation of the FybrikApplication
	DMA dm.FybrikApplication `json:"jsonDMA,omitempty"`

	// Optional message about the action performed
	Message string `json:"message,omitempty"`
}
