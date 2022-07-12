// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"fybrik.io/fybrik/pkg/environment"
	fybrikTLS "fybrik.io/fybrik/pkg/tls"
)

const (
	envOPAServerURL = "OPA_SERVER_URL"
	envServicePort  = "SERVICE_PORT"
)

// NewRouter returns a new router.
func NewRouter(controller *ConnectorController) *gin.Engine {
	router := gin.Default()
	router.POST("/getPoliciesDecisions", controller.GetPoliciesDecisions)
	return router
}

// RootCmd defines the root cli command
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opa-connector",
		Short: "Kubernetes based policy manager connector for Fybrik",
	}
	cmd.AddCommand(RunCmd())
	return cmd
}

// RunCmd defines the command for running the connector
func RunCmd() *cobra.Command {
	ip := ""
	portStr, err := environment.MustGetEnv(envServicePort)
	if err != nil {
		log.Err(err).Msg(envServicePort + " env var is not defined")
		return nil
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Err(err).Msg("error in converting " + envServicePort + " to integer")
		return nil
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run opa connector",
		RunE: func(cmd *cobra.Command, args []string) error {
			gin.SetMode(gin.ReleaseMode)
			// Parse environment variables
			opaServerURL, err := environment.MustGetEnv(envOPAServerURL)
			if err != nil {
				return errors.Wrap(err, "failed to retrieve URL for communicating with OPA server")
			}

			if !strings.HasPrefix(opaServerURL, "https://") && !strings.HasPrefix(opaServerURL, "http://") {
				return errors.New("server URL for OPA server must have http or https schema")
			}

			// Create and start connector
			controller := NewConnectorController(opaServerURL)
			router := NewRouter(controller)
			router.Use(gin.Logger())

			bindAddress := fmt.Sprintf("%s:%d", ip, port)
			if environment.IsUsingTLS() {
				var client kclient.Client
				scheme := runtime.NewScheme()
				err = corev1.AddToScheme(scheme)
				if err != nil {
					return errors.Wrap(err, "unable to add corev1 to schema")
				}
				client, err = kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: scheme})
				if err != nil {
					return errors.Wrap(err, "failed to create a Kubernetes client")
				}

				tlsConfig, err := fybrikTLS.GetServerConfig(&controller.Log, client)
				if err != nil {
					return errors.Wrap(err, "failed to get tls config")
				}
				server := http.Server{Addr: bindAddress, Handler: router, TLSConfig: tlsConfig}
				return server.ListenAndServeTLS("", "")
			}
			controller.Log.Info().Msg(fybrikTLS.TLSDisabledMsg)
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
