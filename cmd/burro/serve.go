package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/marsskom/burro/internal/logger"
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
	handler = loggingMiddleware(handler)

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
			logger.Info("server TLS is enabled with certificates", "cert", serveFlags.Cert, "key", serveFlags.Key)

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
		logger.Info("received signal, shutting down", "signal", sig)

	case err := <-serverErr:
		logger.Error("file server crashed", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	if err != nil {
		logger.Error("file server shutdown failed", "error", err)
	} else {
		logger.Info("file server exited")
	}

	return err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		logger.Info(
			"serve request",
			"remote", r.RemoteAddr,
			"method", r.Method,
			"path", r.URL.Path,
			"user_agent", r.UserAgent(),
			"duration", time.Since(start),
		)
	})
}
