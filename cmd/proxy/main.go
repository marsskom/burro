package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/marsskom/burro/internal/cert"
	"gitlab.com/marsskom/burro/internal/cli"
	"gitlab.com/marsskom/burro/internal/config"
	"gitlab.com/marsskom/burro/internal/database"
	coreLogger "gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence"
	"gitlab.com/marsskom/burro/internal/persistence/repository"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/proxy"

	_ "gitlab.com/marsskom/burro/plugins/registry"
)

func main() {
	// Flags and Config.
	var workspaceFlags config.WorkspaceFlags
	var proxyFlags config.ProxyFlags

	flag.BoolVar(&workspaceFlags.Interactive, "i", false, "interactive flag, false by default")
	flag.StringVar(&workspaceFlags.Workspace, "w", "", "workspace name to load, creates new in memory on empty")
	flag.StringVar(&workspaceFlags.Session, "sess", "", "session ID to load from workspace, creates new in memory on empty")

	flag.IntVar(&proxyFlags.Port, "port", 0, "proxy port override")

	flag.Parse()

	err := config.ValidateWorkspaceFlags(workspaceFlags)
	if err != nil {
		log.Fatal(err)
	}

	path, err := config.ResolvePath("")
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadWithFlags(path, proxyFlags, workspaceFlags)
	if err != nil {
		log.Fatal(err)
	}

	// Logger.
	coreLogger.SetDefault(cfg.Core)

	// Database.
	gooseConfig, err := config.NewGooseConfig()
	if err != nil {
		log.Fatal(err)
	}

	var dbPool *sql.DB
	if workspaceFlags.Workspace != "" {
		db, err := persistence.LoadDatabase(gooseConfig, workspaceFlags.Workspace, "./bin")
		if err != nil {
			log.Fatal(fmt.Errorf("cannot load workspace database: %w", err))
		}

		dbPool = db
	}

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

	// Workspace.
	var workspace *model.Workspace
	if workspaceFlags.Workspace != "" {
		repo := repository.NewWorkspaceRepo(database.New(dbPool))
		workspace, err = repo.LoadWorkspace(context.Background(), workspaceFlags.Workspace)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		workspace = model.NewWorkspace()
	}

	// Session.
	var session *model.Session
	if workspaceFlags.Session != "" {
		repo := repository.NewSessionRepo(database.New(dbPool))
		session, err = repo.LoadSession(context.Background(), workspaceFlags.Session)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		session = model.NewSession()
		workspace.AddSession(session)
	}

	// Proxy.
	px := proxy.NewProxy(pm, session, caCert, caKey)

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

	// Interrupt channel.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-interrupt
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Info("Server shutdown failed", "error", err)
	}

	slog.Info("Server exited")

	if !workspaceFlags.Interactive {
		os.Exit(0)
	}

	if !cli.Confirm("Do you want to save current workspace?", cli.ChoiceNo) {
		os.Exit(0)
	}

	var workspaceName string = workspaceFlags.Workspace
	if workspaceName == "" || !cli.Confirm(fmt.Sprintf("Save session into existing workspace '%s'?", workspaceName), cli.ChoiceYes) {
		workspaceName = cli.AskWithValidator(
			"Enter workspace name (alpha-numeric, _, -):",
			func(input string) error {
				return model.WorkspaceName(input).Validate()
			},
		)
		slog.Info("Workspace name provided by user", "name", workspaceName)
	}
	slog.Debug("Workspace is going to be saved under a name", "name", workspaceName)

	workspace.Name = model.WorkspaceName(workspaceName)

	dbPool, err = persistence.CreateDatabase(gooseConfig, workspaceName, "./bin")
	if err != nil {
		log.Fatal(err)
	}

	workspaceRepo := repository.NewWorkspaceRepo(database.New(dbPool))
	err = workspaceRepo.SaveWorkspace(context.Background(), dbPool, workspace)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
