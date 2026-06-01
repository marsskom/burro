package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/marsskom/burro/internal/cert"
)

var caInitFlags struct {
	DstCert string
	DstKey  string
}

var caGenFlags struct {
	SrcCert string
	SrcKey  string
	DstCert string
	DstKey  string
}

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Manage Burro CA certificates",
}

var certInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate CA pair if it doesn't exist",
	Run: func(cmd *cobra.Command, args []string) {
		createCA()
	},
}

var certGenerateCmd = &cobra.Command{
	Use:   "generate [host]",
	Short: "Generate CA pair for host, localhost by default",
	Run: func(cmd *cobra.Command, args []string) {
		host := "localhost"
		if len(args) == 1 {
			host = strings.TrimSpace(args[0])
		}

		generateCA(host)
	},
}

func init() {
	rootCmd.AddCommand(certCmd)

	certInitCmd.Flags().StringVar(
		&caInitFlags.DstCert,
		"dst-cert",
		"./runtime/certs/ca.pem",
		"save path to tls certificate",
	)
	certInitCmd.Flags().StringVar(
		&caInitFlags.DstKey,
		"dst-key",
		"./runtime/certs/ca.key",
		"save path to tls key",
	)

	certInitCmd.MarkFlagsRequiredTogether("dst-cert", "dst-key")

	certCmd.AddCommand(certInitCmd)

	certGenerateCmd.Flags().StringVar(
		&caGenFlags.SrcCert,
		"src-cert",
		"./runtime/certs/ca.pem",
		"CA certificate",
	)
	certGenerateCmd.Flags().StringVar(
		&caGenFlags.SrcKey,
		"src-key",
		"./runtime/certs/ca.key",
		"CA key",
	)
	certGenerateCmd.Flags().StringVar(
		&caGenFlags.DstCert,
		"dst-cert",
		"./runtime/certs/localhost.pem",
		"save path to tls certificate",
	)
	certGenerateCmd.Flags().StringVar(
		&caGenFlags.DstKey,
		"dst-key",
		"./runtime/certs/localhost.key",
		"save path to tls key",
	)

	certGenerateCmd.MarkFlagsRequiredTogether("src-cert", "src-key")
	certGenerateCmd.MarkFlagRequired("dst-cert")
	certGenerateCmd.MarkFlagRequired("dst-key")

	certCmd.AddCommand(certGenerateCmd)
}

func createCA() {
	err := cert.GenerateCA(
		caInitFlags.DstCert,
		caInitFlags.DstKey,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func generateCA(host string) {
	if host == "" {
		log.Fatal(fmt.Errorf("host value is empty"))
	}

	caCert, caKey, err := cert.LoadCA(caGenFlags.SrcCert, caGenFlags.SrcKey)
	if err != nil {
		log.Fatal(err)
	}

	c, err := cert.GenerateHostCertificate(host, caCert, caKey)
	if err != nil {
		log.Fatal(err)
	}

	err = cert.WriteTLSCertificate(c, caGenFlags.DstCert, caGenFlags.DstKey)
	if err != nil {
		log.Fatal(err)
	}
}
