// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fybrik.io/fybrik/pkg/taxonomy/generator"

	"github.com/spf13/cobra"
)

var (
	taxonomyObjectGenerateInputPath  string
	taxonomyObjectGenerateOutputPath string
	taxonomyObjectTitle              string
)

// generateObjectCmd represents the taxonomy generate command
var generateObjectCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate validation object from CRD YAML files",
	RunE: func(cmd *cobra.Command, args []string) error {
		return generator.GenerateValidationObjectFromCRDs(taxonomyObjectGenerateInputPath, taxonomyObjectGenerateOutputPath, taxonomyObjectTitle)
	},
}

func init() {
	taxonomyCmd.AddCommand(generateObjectCmd)

	generateObjectCmd.Flags().StringVarP(&taxonomyObjectGenerateInputPath, "in", "i", "", "Path to input directory or file")
	_ = generateObjectCmd.MarkFlagRequired("in")
	_ = generateObjectCmd.MarkFlagDirname("in")
	_ = generateObjectCmd.MarkFlagFilename("in", "yaml", "yml")

	generateObjectCmd.Flags().StringVarP(&taxonomyObjectGenerateOutputPath, "out", "o", "", "Path for output directory")
	_ = generateObjectCmd.MarkFlagRequired("out")

	generateObjectCmd.Flags().StringVarP(&taxonomyObjectTitle, "title", "t", "", "Title of the generated object")
	_ = generateObjectCmd.MarkFlagRequired("title")
}
