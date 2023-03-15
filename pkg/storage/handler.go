// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/storagemanager"
	"fybrik.io/fybrik/pkg/storage/registrator"

	// Registration of the implementation agents is done by adding blank imports which invoke init() method of each package
	_ "fybrik.io/fybrik/pkg/storage/impl/mysql"
	_ "fybrik.io/fybrik/pkg/storage/impl/s3"
)

const UnsupportedTypeError string = "unsupported storage type: "

type Handler struct {
	Client kclient.Client
	Log    zerolog.Logger
}

func NewHandler(client kclient.Client) *Handler {
	handler := &Handler{
		Client: client,
		Log:    logging.LogInit(logging.CONNECTOR, "StorageManager"),
	}
	return handler
}

// allocates storage based on the selected storage account by invoking the specific implementation agent
// returns a Connection object in case of success, and an error - otherwise
func (r *Handler) allocateStorage(c *gin.Context) {
	// Parse request
	var request storagemanager.AllocateStorageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.Log.Info().Msg(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error during ShouldBindJSON in allocateStorage "})
		return
	}
	r.Log.Info().Msgf("allocateStorage request for %s", request.AccountType)
	impl, err := registrator.GetAgent(request.AccountType)
	if err != nil {
		r.Log.Info().Msg(err.Error())
		c.JSON(http.StatusNotImplemented, gin.H{"error": UnsupportedTypeError + string(request.AccountType)})
		return
	}
	conn, err := impl.AllocateStorage(&request, r.Client)
	if err != nil {
		r.Log.Info().Msg(err.Error())
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	logging.LogStructure("allocated storage", &conn, &r.Log, zerolog.InfoLevel, false, true)
	c.JSON(http.StatusOK, &storagemanager.AllocateStorageResponse{Connection: &conn})
}

// deletes the existing storage by invoking the specific implementation agent based on the connection type
func (r *Handler) deleteStorage(c *gin.Context) {
	// Parse request
	var request storagemanager.DeleteStorageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.Log.Info().Msg(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error during ShouldBindJSON in deleteStorage"})
		return
	}

	impl, err := registrator.GetAgent(request.Connection.Name)
	if err != nil {
		r.Log.Info().Msg(err.Error())
		c.JSON(http.StatusNotImplemented, gin.H{"error": UnsupportedTypeError + string(request.Connection.Name)})
		return
	}
	r.Log.Info().Msgf("deleteStorage request for %v", request.Connection)
	err = impl.DeleteStorage(&request, r.Client)
	if err != nil {
		r.Log.Info().Msg(err.Error())
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// return a list of supported connection types
func (r *Handler) getSupportedStorageTypes(c *gin.Context) {
	resp := &storagemanager.GetSupportedStorageTypesResponse{ConnectionTypes: registrator.GetRegisteredTypes()}
	r.Log.Info().Msgf("supported connections: %v", resp)
	c.JSON(http.StatusOK, resp)
}
