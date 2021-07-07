// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/mesh-for-data/mesh-for-data/pkg/text"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

//go:embed *.txt
var helpTxtFiles embed.FS

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mesh-for-data",
	Short: "Mesh for Data CLI",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// An optional configuration file to override default flags (available to subcommands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mesh-for-data.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".mesh-for-data" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mesh-for-data")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// loadHelp loads help description from an embedded txt file
func loadHelp(filename string) string {
	txtBytes, err := helpTxtFiles.ReadFile(filename)
	if err != nil {
		// all markdown files are embedded so if this happens then the caller code is wrong
		log.Fatal(err)
	}

	return string(txtBytes)
}

func longHelp(short, filename string) string {
	return fmt.Sprintf(`%s.

Overview:
%s`, short, text.Indent(loadHelp(filename), "  "))
}
