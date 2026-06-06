package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/marsskom/burro/internal/broker"
	"gitlab.com/marsskom/burro/internal/cert"
	"gitlab.com/marsskom/burro/internal/config"
	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/export"
	"gitlab.com/marsskom/burro/internal/grpc"
	"gitlab.com/marsskom/burro/internal/logger"
	coreLogger "gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence"
	"gitlab.com/marsskom/burro/internal/persistence/dbcommand"
	"gitlab.com/marsskom/burro/internal/persistence/repository"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/proxy"

	_ "gitlab.com/marsskom/burro/plugins/registry"
)

var cliFlags config.ProxyFlags

var proxyCmd = &cobra.Command{
	Use:   "proxy [listen]",
	Short: "Start Burro proxy",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cliFlags.Listen = ""
		if len(args) == 1 {
			cliFlags.Listen = strings.TrimSpace(args[0])
		}

		err := run()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	proxyCmd.Flags().BoolVarP(
		&cliFlags.ZeroCfg,
		"z",
		"z",
		false,
		"run proxy without configuration, only logger plugin works, TLS is not mandatory",
	)

	proxyCmd.Flags().StringVarP(
		&cliFlags.GRPCListen,
		"grpc",
		"g",
		"",
		"gRPC listen address",
	)
	proxyCmd.Flags().BoolVar(
		&cliFlags.GRPCDisabled,
		"no-grpc",
		false,
		"gRPC disabled flag",
	)
	proxyCmd.Flags().BoolVar(
		&cliFlags.GRPCDebug,
		"grpc-d",
		false,
		"gRPC debug flag, activates reflection on gRPC server for grpcurl (for example) connections",
	)

	proxyCmd.Flags().StringVarP(
		&cliFlags.WorkDir,
		"workdir",
		"d",
		"",
		"proxy work directory",
	)

	proxyCmd.Flags().StringVarP(
		&cliFlags.Workspace,
		"workspace",
		"w",
		"",
		"workspace to load, creates new in memory on empty",
	)

	proxyCmd.Flags().StringVar(
		&cliFlags.TLSCert,
		"tls-cert",
		"",
		"path to tls certificate",
	)
	proxyCmd.Flags().StringVar(
		&cliFlags.TLSKey,
		"tls-key",
		"",
		"path to tls key",
	)

	proxyCmd.MarkFlagsRequiredTogether("tls-cert", "tls-key")

	proxyCmd.Flags().StringVar(
		&cliFlags.CACert,
		"ca-cert",
		"certs/ca.pem",
		"path to proxy CA certificate in %workdir%",
	)
	proxyCmd.Flags().StringVar(
		&cliFlags.CAKey,
		"ca-key",
		"certs/ca.key",
		"path to proxy CA key in %workdir%",
	)

	proxyCmd.MarkFlagsRequiredTogether("ca-cert", "ca-key")
}

func run() error {
	paths, cfg, err := initConfig(cliFlags)
	if err != nil {
		return err
	}

	coreLogger.SetDefault(verbosity, cfg.Core.LogLevel)

	// Certificates.
	caCertPath := filepath.Join(paths.Home, cliFlags.CACert)
	caKeyPath := filepath.Join(paths.Home, cliFlags.CAKey)
	caCert, caKey, err := cert.LoadCA(caCertPath, caKeyPath)

	if err != nil {
		if !cfg.Proxy.ZeroConfigurationMode {
			return err
		}

		logger.Warn("CA certificates weren't loaded (ignore for zero configuration mode)", "cert", caCertPath, "key", caKeyPath, "err", err)
	} else {
		logger.Info("CA certificates were loaded", "cert", caCertPath, "key", caKeyPath)
	}

	// Broker.
	brokerHub := broker.NewHub()

	// Plugins.
	pm := plugin.NewManager(brokerHub)

	err = plugin.LoadPlugins(paths, cfg, pm)
	if err != nil {
		return err
	}

	workspace, err := initWorkspace(paths, cfg, cliFlags.Workspace)
	if err != nil {
		return err
	}

	// Session.
	session := model.NewSession()
	workspace.AddSession(session)

	// Proxy.
	px := proxy.NewProxy(pm, session, caCert, caKey)

	if cfg.Proxy.Listen == "" {
		return fmt.Errorf("proxy listen address cannot be empty")
	}

	server := &http.Server{
		Addr:    cfg.Proxy.Listen,
		Handler: px,

		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("proxy is listening on", "host", cfg.Proxy.Listen)

	if err := runServer(cfg, brokerHub, server); err != nil {
		return err
	}

	defer func() {
		err := pm.EmitExportPluginsFlush(&export.FileNameVars{
			Session: session.ID,
		})
		if err != nil {
			logger.Error("plugin has failed", "err", err)
		}
	}()

	if cfg.Proxy.ZeroConfigurationMode {
		return nil
	}

	if workspace.GetName() == "" {
		return nil
	}

	err = saveWorkspace(paths, workspace)
	if err != nil {
		return err
	}

	return nil
}

func initConfig(flags config.ProxyFlags) (*config.Paths, *config.Config, error) {
	paths := config.NewPaths(config.ResolveWorkdir(flags.WorkDir))

	if flags.ZeroCfg {
		logger.Warn("proxy runs in zero configuration mode")

		cfg, err := config.NewZeroCfg(flags)
		if err != nil {
			return nil, nil, err
		}

		return paths, cfg, nil
	}

	if _, err := os.Stat(paths.Home); err != nil {
		return nil, nil, err
	}

	cfgPath, err := paths.GetConfigPath("")
	if err != nil {
		return paths, nil, err
	}

	cfg, err := config.LoadWithFlags(cfgPath, flags)
	if err != nil {
		return paths, nil, err
	}

	return paths, cfg, nil
}

func initWorkspace(paths *config.Paths, cfg *config.Config, workspaceName string) (*model.Workspace, error) {
	workspace := model.NewWorkspace(workspaceName)
	if workspace.GetName() == "" {
		return workspace, nil
	}
	if cfg.Proxy.ZeroConfigurationMode {
		return workspace, nil
	}

	if err := workspace.Name.Validate(); err != nil {
		return nil, err
	}

	dbConn := persistence.NewConnection(workspace.GetName(), filepath.Join(paths.Home, "db"))
	if err := dbConn.Open(); err != nil {
		// Ignores if file not found. New db will be created.
		if errors.Is(err, persistence.DBErrorFileNotFound) {
			return workspace, nil
		}

		return nil, err
	}
	defer dbConn.Close()

	repo := repository.NewWorkspaceRepo(database.New(dbConn.DB))
	workspace, err := repo.LoadWorkspace(context.Background(), workspace.GetName())
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

func runServer(cfg *config.Config, hub *broker.Hub, s *http.Server) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)

	// HTTP
	go func() {
		var err error
		if cfg.TLS.Enabled {
			logger.Info("proxy TLS is enabled with certificates", "cert", cfg.TLS.Cert, "key", cfg.TLS.Key)

			err = s.ListenAndServeTLS(cfg.TLS.Cert, cfg.TLS.Key)
		} else {
			err = s.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// GRPC
	gRPCWrapper := grpc.NewServerWrapper(cfg, hub)
	gRPCWrapper.Start(serverErr)

	select {
	case sig := <-interrupt:
		logger.Info("received signal, shutting down", "signal", sig)

	case err := <-serverErr:
		logger.Error("server crashed", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	if err != nil {
		logger.Error("HTTP server shutdown failed", "error", err)
	} else {
		logger.Info("HTTP server exited")
	}

	gRPCWrapper.Stop(ctx)

	return err
}

func saveWorkspace(paths *config.Paths, w *model.Workspace) error {
	if err := w.Name.Validate(); err != nil {
		return fmt.Errorf("cannot save workspace with invali name: %w", err)
	}

	dbConn := persistence.NewConnection(w.GetName(), filepath.Join(paths.Home, "db"))
	err := dbConn.OpenOrCreate()
	if err != nil {
		return err
	}
	defer dbConn.Close()

	logger.Debug("workspace is going to be saved under a name", "name", w.GetName())

	err = dbcommand.UpsertWorkspaceCommand(context.Background(), dbConn.DB, w)
	if err != nil {
		return err
	}

	return nil
}
