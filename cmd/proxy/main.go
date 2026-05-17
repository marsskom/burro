package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/lukechampine/freeze"
	"gitlab.com/marsskom/burro/internal/config"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/proxy"
	"gitlab.com/marsskom/burro/plugins/logger"
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

	pm := plugin.NewManager()

	pm.Register(logger.New())

	px := proxy.New(pm)

	log.Printf("proxy listening on :%d\n", cfg.Proxy.Port)

	err = http.ListenAndServe(
		fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port),
		px,
	)
	if err != nil {
		panic(err)
	}
}
