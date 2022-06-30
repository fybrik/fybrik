// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"emperror.dev/errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"fybrik.io/fybrik/connectors/opa/utils"
	"fybrik.io/fybrik/pkg/environment"
	fybrikTLS "fybrik.io/fybrik/pkg/tls"
)

const (
	envOPAServerURL = "OPA_SERVER_URL"
	commandPort     = 8080
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
	port := commandPort
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
			if utils.GetTLSEnabled() {
				controller.Log.Info().Msg("TLS is enabled")
				scheme := runtime.NewScheme()
				err = corev1.AddToScheme(scheme)
				if err != nil {
					return errors.Wrap(err, "unable to add corev1 to schema")
				}
				client, err := kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: scheme})
				if err != nil {
					return errors.Wrap(err, "failed to create a Kubernetes client")
				}
				config, err := fybrikTLS.GetServerTLSConfig(&controller.Log, client, utils.GetCertSecretName(), utils.GetCertSecretNamespace(),
					utils.GetCACERTSecretName(), utils.GetCACERTSecretNamespace(), utils.GetMTLSEnabled())
				if err != nil {
					return nil
				}

				server := http.Server{Addr: bindAddress, Handler: router, TLSConfig: config}
				return server.ListenAndServeTLS("", "")
			} else {
				controller.Log.Info().Msg("TLS is disabled")
				return router.Run(bindAddress)
			}
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
