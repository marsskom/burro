package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"gitlab.com/marsskom/burro/internal/cert"
	"gitlab.com/marsskom/burro/internal/config"
	coreLogger "gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/proxy"

	_ "gitlab.com/marsskom/burro/plugins/registry"
)

func main() {
	// Flags and Config.
	var flags config.ProxyFlags

	flag.IntVar(&flags.Port, "port", 0, "proxy port override")

	flag.Parse()

	path, err := config.ResolvePath("")
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadWithFlags(path, flags)
	if err != nil {
		log.Fatal(err)
	}

	// Logger.
	coreLogger.SetDefault(cfg.Core)

	// Certificates.
	caCert, caKey, err := cert.LoadCA(
		"./certs/ca.pem",
		"./certs/ca.key",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Plugins.
	pm := plugin.NewManager()

	err = plugin.LoadPlugins(cfg, pm)
	if err != nil {
		log.Fatal(err)
	}

	// Proxy.
	px := proxy.New(pm, caCert, caKey)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port),
		Handler: px,

		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("Proxy is listening on host", "host", cfg.Proxy.Host)
	slog.Info("Proxy is listening on port", "port", cfg.Proxy.Port)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
