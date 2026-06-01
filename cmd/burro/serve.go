package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve [address] [directory]",
	Short: "Run simple static file server on directory",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatal("missing arguments")
		}

		addr := strings.TrimSpace(args[0])
		dir := strings.TrimSpace(args[1])

		err := serve(addr, dir)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var serveFlags struct {
	Cert string
	Key  string
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(
		&serveFlags.Cert,
		"cert",
		"c",
		"",
		"path to tls certificate",
	)
	serveCmd.Flags().StringVarP(
		&serveFlags.Key,
		"key",
		"k",
		"",
		"path to tls key",
	)

	serveCmd.MarkFlagsRequiredTogether("cert", "key")
}

func serve(addr, dir string) error {
	handler := http.FileServer(http.Dir(dir))

	s := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)

	go func() {
		var err error
		if serveFlags.Cert != "" {
			slog.Info("server TLS is enabled with certificates", "cert", serveFlags.Cert, "key", serveFlags.Key)

			err = s.ListenAndServeTLS(serveFlags.Cert, serveFlags.Key)
		} else {
			err = s.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case sig := <-interrupt:
		slog.Info("received signal, shutting down", "signal", sig)

	case err := <-serverErr:
		slog.Error("file server crashed", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	if err != nil {
		slog.Error("file server shutdown failed", "error", err)
	} else {
		slog.Info("file server exited")
	}

	return err
}
