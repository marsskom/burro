package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "burro",
	Short:   "Burro proxy and security inspection tool",
	Version: "v0.2.1",
}

var verbosity int

func init() {
	rootCmd.PersistentFlags().CountVarP(
		&verbosity,
		"verbose",
		"v",
		"increase verbosity (-v, -vv, -vvv)",
	)

	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
