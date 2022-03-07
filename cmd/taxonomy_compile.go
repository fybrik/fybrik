// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"github.com/spf13/cobra"

	"fybrik.io/fybrik/pkg/taxonomy/compile"
	taxonomyio "fybrik.io/fybrik/pkg/taxonomy/io"
)

const (
	baseLiteral   = "base"
	outLiteral    = "out"
	yamlExtension = "yaml"
	ymlExtension  = "yml"
	jsonExtension = "json"
)

var (
	taxonomyCompileBasePath string
	taxonomyCompileOutPath  string
	taxonomyCompileCodegen  bool
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile --out <outputFile> --base <baseFile> [<layerFile> ...] [--codegen]",
	Short: "Generate a taxonomy.json file from base taxonomy and taxonomy layers",
	Args:  cobra.ArbitraryArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{yamlExtension, ymlExtension, jsonExtension}, cobra.ShellCompDirectiveDefault
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := compile.Files(taxonomyCompileBasePath, args, compile.WithCodeGenerationTarget(taxonomyCompileCodegen))
		if err != nil {
			return err
		}
		return taxonomyio.WriteDocumentToFile(result, taxonomyCompileOutPath)
	},
	DisableFlagsInUseLine: true,
}

func init() {
	taxonomyCmd.AddCommand(compileCmd)

	compileCmd.Flags().StringVarP(&taxonomyCompileBasePath, baseLiteral, "b", "", "File with base taxonomy definitions (required)")
	_ = compileCmd.MarkFlagFilename(baseLiteral, yamlExtension, ymlExtension, jsonExtension)
	_ = compileCmd.MarkFlagRequired(baseLiteral)

	compileCmd.Flags().StringVarP(&taxonomyCompileOutPath, outLiteral, "o", "taxonomy.json", "Path for output file")
	_ = compileCmd.MarkFlagFilename(outLiteral, yamlExtension, ymlExtension, jsonExtension)

	compileCmd.Flags().BoolVar(&taxonomyCompileCodegen, "codegen", false,
		"Best effort to make output suitable for code generation tools")
}
