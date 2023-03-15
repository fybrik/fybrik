// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"

	"emperror.dev/errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"fybrik.io/fybrik/pkg/environment"
)

const ServerPortKey string = "SERVER_PORT"

// NewRouter returns a new router.
func NewRouter(handler *Handler) *gin.Engine {
	router := gin.Default()
	router.POST("/allocateStorage", handler.allocateStorage)
	router.DELETE("/deleteStorage", handler.deleteStorage)
	router.GET("/getSupportedStorageTypes", handler.getSupportedStorageTypes)
	return router
}

// RootCmd defines the root cli command
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "Storage Manager",
		Short: "Built-in storage manager for Fybrik",
	}
	cmd.AddCommand(RunCmd())
	return cmd
}

// RunCmd defines the command for running the storage manager
func RunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run storage manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			gin.SetMode(gin.ReleaseMode)
			client, err := K8sInit()
			if err != nil {
				return errors.Wrap(err, "failed to create a Kubernetes client")
			}
			handler := NewHandler(client)
			router := NewRouter(handler)
			router.Use(gin.Logger())
			port, err := environment.MustGetEnv(ServerPortKey)
			if err != nil {
				handler.Log.Err(err).Msg(ServerPortKey + " env var is not defined")
				return err
			}
			bindAddress := ":" + port
			return router.Run(bindAddress)
		},
	}
	return cmd
}

func main() {
	// Run the cli
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
