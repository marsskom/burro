package main

import (
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"gitlab.com/marsskom/burro/internal/cert"
	"gitlab.com/marsskom/burro/internal/config"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Manage Burro CA certificates",
}

var certInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate CA pair if it doesn't exist",
	Run: func(cmd *cobra.Command, args []string) {
		initCA()
	},
}

func initCA() {
	paths := config.NewPaths(config.ResolveHome(""))

	err := cert.GenerateCA(
		filepath.Join(paths.Home, "certs/ca.pem"),
		filepath.Join(paths.Home, "certs/ca.key"),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(certCmd)

	certCmd.AddCommand(certInitCmd)
}
