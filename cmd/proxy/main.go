package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
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
	"gitlab.com/marsskom/burro/internal/export"
	coreLogger "gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence"
	"gitlab.com/marsskom/burro/internal/persistence/dbcommand"
	"gitlab.com/marsskom/burro/internal/persistence/repository"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/proxy"

	_ "gitlab.com/marsskom/burro/plugins/registry"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Error has occured", "err", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	cfg, wf, err := initConfig()
	if err != nil {
		return err
	}

	coreLogger.SetDefault(cfg.Core)

	// Goose Config.
	gc, err := config.NewGooseConfig()
	if err != nil {
		return err
	}

	// Certificates.
	caCert, caKey, err := cert.LoadCA(
		"./certs/ca.pem",
		"./certs/ca.key",
	)
	if err != nil {
		return err
	}

	// Plugins.
	pm := plugin.NewManager()

	err = plugin.LoadPlugins(cfg, pm)
	if err != nil {
		return err
	}

	workspace, err := initWorkspace(wf, gc)
	if err != nil {
		return err
	}

	// Session.
	session := model.NewSession()
	workspace.AddSession(session)

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

	if err := runServer(server); err != nil {
		return err
	}

	defer func() {
		err := pm.EmitExportPluginsFlush(&export.FileNameVars{
			Session: session.ID,
		})
		if err != nil {
			slog.Error("plugin has failed", "err", err)
		}
	}()

	if !wf.Interactive {
		return nil
	}

	if workspace.GetName() == "" {
		okSaveWorkspace, err := cli.Confirm(
			cli.IO{
				In:  bufio.NewReader(os.Stdin),
				Out: os.Stdout,
			},
			"Do you want to save current workspace?",
			cli.ChoiceNo,
		)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if !okSaveWorkspace {
			return nil
		}
	}

	if workspace.GetName() != "" {
		okSaveWorkspace, err := cli.Confirm(
			cli.IO{
				In:  bufio.NewReader(os.Stdin),
				Out: os.Stdout,
			},
			fmt.Sprintf("Add session to existing workspace [%s]?", workspace.GetName()),
			cli.ChoiceYes,
		)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if !okSaveWorkspace {
			return nil
		}
	}

	err = saveWorkspace(workspace, gc)
	if err != nil {
		return err
	}

	return nil
}

func initConfig() (*config.Config, config.WorkspaceFlags, error) {
	var workspaceFlags config.WorkspaceFlags
	var proxyFlags config.ProxyFlags

	flag.BoolVar(&workspaceFlags.Interactive, "i", false, "interactive flag, false by default")
	flag.StringVar(&workspaceFlags.Workspace, "w", "", "workspace name to load, creates new in memory on empty")

	flag.IntVar(&proxyFlags.Port, "port", 0, "proxy port override")

	flag.Parse()

	path, err := config.ResolvePath("")
	if err != nil {
		return nil, config.WorkspaceFlags{}, err
	}

	cfg, err := config.LoadWithFlags(path, proxyFlags)
	if err != nil {
		return nil, config.WorkspaceFlags{}, err
	}

	return cfg, workspaceFlags, nil
}

func initWorkspace(
	wf config.WorkspaceFlags,
	gc *config.GooseConfig,
) (*model.Workspace, error) {
	workspace := model.NewWorkspace(wf.Workspace)
	if wf.Workspace == "" {
		return workspace, nil
	}

	if err := workspace.Name.Validate(); err != nil {
		return nil, err
	}

	dbConnection := persistence.NewConnection(gc, workspace.GetName(), "./bin")
	if err := dbConnection.Open(); err != nil {
		return nil, err
	}
	defer dbConnection.Close()

	repo := repository.NewWorkspaceRepo(database.New(dbConnection.DB))
	workspace, err := repo.LoadWorkspace(context.Background(), workspace.GetName())
	if err != nil {
		dbConnection.Close()

		return nil, err
	}

	return workspace, nil
}

func runServer(s *http.Server) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)

	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case sig := <-interrupt:
		slog.Info("Received signal, shutting down", "signal", sig)

	case err := <-serverErr:
		slog.Error("Server crashed", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	if err != nil {
		slog.Error("Server shutdown failed", "error", err)
	} else {
		slog.Info("Server exited")
	}

	return err
}

func saveWorkspace(w *model.Workspace, gc *config.GooseConfig) error {
	if w.GetName() == "" {
		inputName, err := cli.AskWithValidator(
			cli.IO{
				In:  bufio.NewReader(os.Stdin),
				Out: os.Stdout,
			},
			"Enter workspace name (alpha-numeric, _, -):",
			func(input string) error {
				return model.WorkspaceName(input).Validate()
			},
		)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		slog.Info("Workspace name provided by user", "name", inputName)

		if w.GetName() != inputName {
			w.Name = model.WorkspaceName(inputName)
		}
	}

	dbConn := persistence.NewConnection(gc, w.GetName(), "./bin")
	err := dbConn.Open()
	if err != nil {
		if errors.Is(err, persistence.DBErrorFileNotFound) {
			if err := dbConn.Create(); err != nil {
				return fmt.Errorf("create db: %w", err)
			}
		} else {
			return err
		}
	}

	defer dbConn.Close()

	slog.Debug("Workspace is going to be saved under a name", "name", w.GetName())

	err = dbcommand.TransactionalSaveWorkspace(context.Background(), dbConn.DB, w)
	if err != nil {
		return err
	}

	return nil
}
