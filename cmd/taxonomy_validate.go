// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package cmd

/*
import (
	"fmt"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"

	"fybrik.io/fybrik/pkg/taxonomy/validate"
)

// define the "validate" command" to run taxonomy.ValidateSchema
var validateCmd = &cobra.Command{
	Use:           "validate FILE",
	Short:         "validates a taxonomy JSON schema",
	Args:          require.ExactArgs(1),
	SilenceErrors: true,
	SilenceUsage:  true,
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

		if err := validate.IsDraft4(filename); err != nil {
			return errors.Wrapf(err, "%s file is not a valid draft4 JSON schema", filename)
		}
		if err := validate.IsStructuralSchema(filename); err != nil {
			return errors.Wrapf(err, "%s file is not a valid structural schema", filename)
		}
		fmt.Printf("%s validated successfully\n", filename)
		return nil
	},
}

func init() {
	taxonomyCmd.AddCommand(validateCmd)
}
*/
