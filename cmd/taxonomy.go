// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

// taxonomyCmd represents the taxonomy command
var taxonomyCmd = &cobra.Command{
	Use:   "taxonomy",
	Short: "Commands for working with taxonomy definitions",
}

func init() {
	rootCmd.AddCommand(taxonomyCmd)
}
