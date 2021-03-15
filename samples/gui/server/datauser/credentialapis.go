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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var k8sClient *K8sClient

// UserCredentials contains the credentials needed to access a given system for the purpose of running a specific compute function.
type UserCredentials struct {
	SecretName  string                 `json:"secretName"`
	System      string                 `json:"system"`      // system to access using the credentials, e.g. Egeria
	Credentials map[string]interface{} `json:"credentials"` // often username and password, but could be token or other types of credentials
}

// CredentialRoutes is a list of the REST APIs supported by the backend of the Data User GUI
func CredentialRoutes(client *K8sClient) *chi.Mux {
	k8sClient = client // global variable used by all funcs in this package

	router := chi.NewRouter()
	router.Get("/{secret}", GetCredentials)
	router.Delete("/{secret}", DeleteCredentials)
	router.Post("/", StoreCredentials)
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
	if k8sClient == nil {
		err := render.Render(w, r, ErrConfigProblem(errors.New("No k8sClient set")))
		if err != nil {
			log.Printf(err.Error() + " upon No k8sClient set")
		}
	}
	secretName := chi.URLParam(r, "secret")

	// Call kubernetes to get the M4DApplication CRD
	secret, err := k8sClient.GetSecret(secretName)
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}
	render.JSON(w, r, secret.Data) // Return the credentials as json
}

// DeleteCredentials deletes the secret
func DeleteCredentials(w http.ResponseWriter, r *http.Request) {
	log.Println("In DeleteCredentials")
	if k8sClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No client set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon no client set")
		}
	}

	secretName := chi.URLParam(r, "secret")

	// Call kubernetes to get the M4DApplication CRD
	err := k8sClient.DeleteSecret(secretName, nil)
	if err != nil {
		suberr := render.Render(w, r, ErrRender(err))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon " + err.Error())
		}
		return
	}

	render.Status(r, http.StatusOK)
	result := CredsSuccessResponse{Name: secretName, Message: "Deleted!!"}
	render.JSON(w, r, result)
}

// StoreCredentials stores the credentials
func StoreCredentials(w http.ResponseWriter, r *http.Request) {
	var err error

	log.Println("In StoreCredentials")
	if k8sClient == nil {
		suberr := render.Render(w, r, ErrConfigProblem(errors.New("No k8sClient set")))
		if suberr != nil {
			log.Printf(suberr.Error() + " upon No k8sClient set")
		}
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var userCredentials UserCredentials

	// Create the golang structure from the json
	err = decoder.Decode(&userCredentials)
	if err != nil {
		log.Print("err = " + err.Error())
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Create a secret
	secretStruct := v1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      userCredentials.SecretName,
		Namespace: k8sClient.namespace,
	},
		Data: map[string][]byte{},
		Type: "Opaque",
	}

	// add system name as prefix to the credentials
	for key, val := range userCredentials.Credentials {
		bytes, err := json.Marshal(val)
		if err != nil {
			log.Print("err = " + err.Error())
			_ = render.Render(w, r, ErrConfigProblem(err))
			return
		}
		secretStruct.Data[userCredentials.System+"_"+key] = bytes
	}

	secret, err := k8sClient.CreateOrUpdateSecret(&secretStruct)
	if err != nil {
		log.Print("err = " + err.Error())
		_ = render.Render(w, r, ErrConfigProblem(err))
		return
	}

	// Return the results
	render.Status(r, http.StatusCreated)
	result := CredsSuccessResponse{Name: secret.Name, Message: "Created!!"}
	render.JSON(w, r, result)
}

// ---------------- Responses -----------------------------------------

// CredsSuccessResponse - Structure returned when REST API is successful
type CredsSuccessResponse struct {
	// JSON representation of the Secret
	Secret string `json:"jsonDMA,omitempty"`

	// Secret name
	Name string `json:"name,omitempty"`

	// Optional message about the action performed
	Message string `json:"message,omitempty"`
}
