// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"fybrik.io/fybrik/pkg/taxonomy"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
)

// define the "validate" command" to run taxonomy.ValidateSchema
var validateCmd = &cobra.Command{
	Use:   "validate FILE",
	Short: "validates a taxonomy JSON schema",
	Args:  require.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// Allow file completion when completing the argument for the name
			// which could be a path
			return []string{"json"}, cobra.ShellCompDirectiveDefault
		}
		// No more completions, so disable file completion
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		err := taxonomy.ValidateSchema(filename)
		if err == nil {
			fmt.Println("Validate: " + filename + " successfully")
		}
		return err
	},
}

func init() {
	taxonomyCmd.AddCommand(validateCmd)
}
