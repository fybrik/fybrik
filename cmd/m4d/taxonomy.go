package main

import (
	"fmt"
	"strings"

	"github.com/mesh-for-data/mesh-for-data/pkg/taxonomy"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
)

// Add root "taxonomy" command option
func newTaxonomyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "taxonomy",
		Short: "tools for working with taxonomies",
		Args:  require.NoArgs,
	}
	cmd.AddCommand(newTaxonomySchemaCmd())
	return cmd
}

// Add "schema" command option as part of the "taxonomy" commands
func newTaxonomySchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "commands for working with taxonomy JSON schema files",
		Args:  require.NoArgs,
	}
	cmd.AddCommand(newTaxonomyValidateCmd())
	return cmd
}

// Add  "validate" schema file command under the schema command
func newTaxonomyValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate FILE",
		Short: "validates a taxonomy JSON schema file",
		Args:  require.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				// Allow file completion when completing the argument for the name
				// which could be a path
				return nil, cobra.ShellCompDirectiveDefault
			}
			// No more completions, so disable file completion
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := taxonomy.ValidateSchema(args[0])
			if err == nil {
				fmt.Println("Validate: " + strings.Join(args, " ") + " successfully")
			}
			return err
		},
	}

	return cmd
}
