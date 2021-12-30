// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"

	"emperror.dev/errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	"fybrik.io/fybrik/connectors/katalog/pkg/connector"
)

// RootCmd defines the root cli command
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "katalog",
		Short: "Kubernetes based data catalog for Fybrik",
	}
	cmd.AddCommand(RunCmd())
	return cmd
}

// RunCmd defines the command for running the connector
func RunCmd() *cobra.Command {
	ip := ""
	port := 8080
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the connector",
		RunE: func(cmd *cobra.Command, args []string) error {
			gin.SetMode(gin.ReleaseMode)

			scheme := runtime.NewScheme()
			err := v1alpha1.AddToScheme(scheme)
			if err != nil {
				return errors.Wrap(err, "unable to add katalog v1alpha1 to schema")
			}

			client, err := kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: scheme})
			if err != nil {
				return errors.Wrap(err, "failed to create a Kubernetes client")
			}

			handler := connector.NewHandler(client)
			router := connector.NewRouter(handler)
			router.Use(gin.Logger())

			bindAddress := fmt.Sprintf("%s:%d", ip, port)
			return router.Run(bindAddress)
		},
	}
	cmd.Flags().StringVar(&ip, "ip", ip, "IP address")
	cmd.Flags().IntVar(&port, "port", port, "Listening port")
	return cmd
}

func main() {
	// Run the cli
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
