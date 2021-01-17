// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ibm/the-mesh-for-data/connectors/katalog/pkg/connector"
)

//go:generate openapi2crd manifests/spec.yaml -g katalog.m4d.ibm.com/v1alpha1/Asset -o install/crds.gen.yaml
//go:generate oapi-codegen -generate "types,skip-prune" -package taxonomy -o pkg/taxonomy/taxonomy.gen.go ./manifests/taxonomy.yaml
//go:generate oapi-codegen -generate "types,skip-prune" -import-mapping=taxonomy.yaml:github.com/ibm/the-mesh-for-data/connectors/katalog/pkg/taxonomy -package api -o pkg/api/spec.gen.go ./manifests/spec.yaml
//go:generate crdoc --template ./docs/main.tmpl --resources ./install/ --output ./docs/README.md

// RootCmd defines the root cli command
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "katalog",
		Short: "Kubernetes based data catalog for Mesh for Data",
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
			address := fmt.Sprintf("%s:%d", ip, port)
			return connector.Start(address)
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
