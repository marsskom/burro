package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "burro",
	Short: "Burro proxy and security inspection tool",
}

var verbosity int

func init() {
	rootCmd.PersistentFlags().CountVarP(
		&verbosity,
		"verbose",
		"v",
		"increase verbosity (-v, -vv, -vvv)",
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
