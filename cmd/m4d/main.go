package main // import "github.com/ibm/the-mesh-for-data/cmd/m4d"

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "m4d",
		Short:         "Mesh for Data CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlags(cmd.Flags())
		},
	}

	// Add subcommands
	cmd.AddCommand(
		newTaxonomyCmd(),
	)

	cobra.OnInitialize(initConfig)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func initConfig() {
	viper.AutomaticEnv()
}

func main() {
	// Run the cli
	if err := rootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
