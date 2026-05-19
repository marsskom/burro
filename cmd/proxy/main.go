package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/lukechampine/freeze"
	"gitlab.com/marsskom/burro/internal/config"
	coreLogger "gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/proxy"

	_ "gitlab.com/marsskom/burro/plugins/registry"
)

func main() {
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

	cfg = freeze.Object(cfg).(*config.Config)

	coreLogger.SetDefault(cfg.Core)

	pm := plugin.NewManager()

	err = plugin.LoadPlugins(cfg, pm)
	if err != nil {
		log.Fatal(err)
	}

	px := proxy.New(pm)

	slog.Info("Proxy is listening on host", "host", cfg.Proxy.Host)
	slog.Info("Proxy is listening on port", "port", cfg.Proxy.Port)

	err = http.ListenAndServe(
		fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port),
		px,
	)
	if err != nil {
		panic(err)
	}
}
