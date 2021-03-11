// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"github.com/ibm/the-mesh-for-data/pkg/vault"
)

var k8sClient *DMAClient
var vaultConnection vault.Interface

// UserCredentials contains the credentials needed to access a given system for the purpose of running a specific compute function.
type UserCredentials struct {
	System           string                 `json:"system"`
	M4DApplicationID string                 `json:"m4dapplicationID"`
	Credentials      map[string]interface{} `json:"credentials"` // often username and password, but could be token or other types of credentials
}

// CredentialRoutes is a list of the REST APIs supported by the backend of the Data User GUI
func CredentialRoutes(client *DMAClient) *chi.Mux {
	k8sClient = client // global variable used by all funcs in this package

	router := chi.NewRouter()
	router.Get("/{namespace}/{m4dapplicationID}/{system}", GetCredentials)
	router.Delete("/{namespace}/{m4dapplicationID}/{system}", DeleteCredentials)
	router.Post("/", CreateCredentials)
	router.Put("/", UpdateCredentials)
	router.Options("/*", CredentialOptions)
	return router
}

// CredentialOptions returns an OK status, but more importantly its header is set to indicate
// that future POST, PUT and DELETE calls are allowed as per the header values set when the router was initiated in main.go
func CredentialOptions(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
}

// GetCredentials returns the credentials for a specified system, namespace and compute
func GetCredentials(w http.ResponseWriter, r *http.Request) {
	log.Println("In GetCredentials")

	var err error

	if k8sClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No k8sClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon No k8sClient set")
		}
	}
	if vaultConnection == nil {
		vaultConnection, err = initVault()
		if err != nil {
			suberr := render.Render(w, r, ErrConfigProblem(errors.New("No vault client set")))
			if suberr != nil {
				log.Printf(suberr.Error() + "upon no vault client set")
			}
		}
	}

	// Call vault to get the credentials
	creds, err2 := vaultConnection.GetSecret(utils.GenerateUserCredentialsSecretName(chi.URLParam(r, "namespace"), chi.URLParam(r, "m4dapplicationID"), chi.URLParam(r, "system")))
	if err2 != nil {
		suberr := render.Render(w, r, SysErrRender(err2))
		if suberr != nil {
			log.Printf(suberr.Error() + "upon no vault client set")
		}
		return
	}

	render.JSON(w, r, creds) // Return the M4DApplication as json
}

// UpdateCredentials calls CreateCredentials because we are using kv-v1, which overwrites credentials so no need to handle updates
func UpdateCredentials(w http.ResponseWriter, r *http.Request) {
	log.Println("In UpdateM4DApplication")

	CreateCredentials(w, r)
}

// DeleteCredentials deletes the credentials stored in the indicated vaultPath
func DeleteCredentials(w http.ResponseWriter, r *http.Request) {
	log.Println("In DeleteCredentials")

	// Call vault to delete the user credentials
	var err error
	if vaultConnection == nil {
		vaultConnection, err = initVault()
		if err != nil {
			suberr := render.Render(w, r, ErrConfigProblem(errors.New("No vault client set")))
			if suberr != nil {
				log.Printf(suberr.Error() + "upon no vault client set")
			}
		}
	}

	err2 := vaultConnection.DeleteSecret(utils.GenerateUserCredentialsSecretName(chi.URLParam(r, "namespace"), chi.URLParam(r, "m4dapplicationID"), chi.URLParam(r, "system")))
	if err2 != nil {
		suberr := render.Render(w, r, ErrConfigProblem(err2))
		if suberr != nil {
			log.Printf(suberr.Error() + "upon " + err2.Error())
		}
	}

	render.Status(r, http.StatusOK)
	result := CredsSuccessResponse{Message: "Deleted!!"}
	render.JSON(w, r, result)
}

// CreateCredentials stores the credentials for the indicated system and m4dapplication name in vault and returns the vaultPath to which they were written
// The vault path created includes the namespace in which this service is running, since the m4d control plane services the entire cluster.
func CreateCredentials(w http.ResponseWriter, r *http.Request) {
	var err error

	log.Println("In CreateCredentials")
	if k8sClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No k8sClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon No k8sClient set")
		}
	}

	if vaultConnection == nil {
		vaultConnection, err = initVault()
		if err != nil {
			suberr := render.Render(w, r, ErrConfigProblem(errors.New("No vault client set")))
			if suberr != nil {
				log.Printf(suberr.Error() + "upon no vault client set")
			}
		}
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var userCredentials UserCredentials

	// Create the golang structure from the json
	err = decoder.Decode(&userCredentials)
	if err != nil {
		suberr := render.Render(w, r, ErrInvalidRequest(err))
		if suberr != nil {
			log.Printf(suberr.Error() + "upon " + err.Error())
		}
		return
	}

	// Write the credentials to vault
	vaultPath := utils.GenerateUserCredentialsSecretName(k8sClient.namespace, userCredentials.M4DApplicationID, userCredentials.System)
	err2 := vaultConnection.AddSecret(vaultPath, userCredentials.Credentials)
	log.Printf("vaultPath = " + vaultPath)
	if err2 != nil {
		log.Print("err = " + err.Error())
		suberr := render.Render(w, r, ErrConfigProblem(err2))
		if suberr != nil {
			log.Printf(suberr.Error() + "upon " + err.Error())
		}
		return
	}

	// Return the results
	render.Status(r, http.StatusCreated)
	result := CredsSuccessResponse{VaultPath: vaultPath, Message: "Created!!"}
	render.JSON(w, r, result)
}

func initVault() (vault.Interface, error) {
	vaultConnection, err := vault.InitConnection(utils.GetVaultAddress(), utils.GetVaultToken())
	if err != nil {
		return nil, err
	}
	return vaultConnection, nil
}

// ---------------- Responses -----------------------------------------

// CredsSuccessResponse - Structure returned when REST API is successful
type CredsSuccessResponse struct {

	// VaultPath for credentials created, updated, deleted
	VaultPath string `json:"vaultPath,omitempty"`

	// Optional message about the action performed
	Message string `json:"message,omitempty"`
}
