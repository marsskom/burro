package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/marsskom/burro/internal/config"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init <path> [name]",
	Short: "Init app workspace directory under %name% (default runtime) in %path%",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("path is required")
		}

		path := strings.TrimSpace(args[0])

		name := "runtime"
		if len(args) > 1 {
			name = strings.TrimSpace(args[1])
		}

		err := cmdInit(path, name)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var initFlags struct {
	DryRun bool
	Force  bool
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(
		&initFlags.DryRun,
		"dry-run",
		false,
		"only prints the structure will be generated",
	)
	initCmd.Flags().BoolVar(
		&initFlags.Force,
		"force",
		false,
		"overwrites files with default values even the files exist",
	)
}

func cmdInit(path, name string) error {
	if path == "" || name == "" {
		return fmt.Errorf("path and name arguments are required")
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot conver path '%s' into absolute: %w", path, err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path doesn't exist: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("path '%s' is not a directory", path)
	}

	appDir := filepath.Join(path, name)

	dirs := []string{
		appDir,
		filepath.Join(appDir, "artifacts"),
		filepath.Join(appDir, "certs"),
		filepath.Join(appDir, "db"),
		filepath.Join(appDir, "plugins"),
		filepath.Join(appDir, "plugins", "policy"),
		filepath.Join(appDir, "plugins", "policy", "data"),
		filepath.Join(appDir, "plugins", "policy", "actions"),
		filepath.Join(appDir, "plugins", "luaplugin"),
		filepath.Join(appDir, "plugins", "luaplugin", "scripts"),
		filepath.Join(appDir, "plugins", "luaplugin", "scripts", "metric"),
	}

	fmt.Println("Directories:")
	for _, dir := range dirs {
		if initFlags.DryRun {
			fmt.Printf(" - dir '%s' would be created\n", dir)
			continue
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create '%s' directory: %w", dir, err)
		}

		fmt.Printf(" - '%s' created\n", dir)
	}

	fmt.Println("")
	fmt.Println("Configs:")

	appConfig := filepath.Join(appDir, "config.yml")
	if err = write(appConfig, func() error {
		return writeAppConfig(appConfig)
	}); err != nil {
		return err
	}

	policyConfig := filepath.Join(appDir, "plugins", "policy", "config.yml")
	if err = write(policyConfig, func() error {
		return writePolicyConfig(policyConfig)
	}); err != nil {
		return err
	}

	whitelist := filepath.Join(appDir, "plugins", "policy", "data", "whitelist.txt")
	if err = write(whitelist, func() error {
		return os.WriteFile(whitelist, []byte{}, 0644)
	}); err != nil {
		return err
	}

	blacklist := filepath.Join(appDir, "plugins", "policy", "data", "blacklist.txt")
	if err = write(blacklist, func() error {
		return os.WriteFile(blacklist, []byte{}, 0644)
	}); err != nil {
		return err
	}

	return nil
}

func write(filename string, fn func() error) error {
	if initFlags.DryRun {
		fmt.Printf(" - config '%s' would be created\n", filename)
		return nil
	}

	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) || initFlags.Force {
		if err := fn(); err != nil {
			return fmt.Errorf("cannot create config '%s': %w", filename, err)
		}

		fmt.Printf(" - config '%s' created\n", filename)
	}

	return nil
}

func writeAppConfig(path string) error {
	cfg := config.Config{
		Version: 1,
		Core: config.CoreConfig{
			LogLevel: "error",
			Plugins: config.CorePluginsConfig{
				Dir:    "plugins",
				Config: "config.yml",
			},
		},
		Proxy: config.ProxyConfig{
			ZeroConfigurationMode: true,
			Listen:                "localhost:8080",
		},
		GRPC: config.GRPCConfig{
			Enabled: true,
			Debug:   false,
			Listen:  "localhost:7777",
		},
		TLS: config.TLSConfig{
			Enabled:  false,
			Insecure: false,
			Cert:     "",
			Key:      "",
		},
		Plugins: map[string]any{
			"logger": map[string]any{},
			"policy": map[string]any{},
			"harexport": map[string]any{
				"enabled":  false,
				"file":     "%session-%datetime%.har",
				"override": true,
			},
			"luaplugin": map[string]any{
				"enabled":  false,
				"priority": 90,
				"dir":      "scripts",
				"scripts": map[string]any{
					"metric": map[string]any{
						"enabled":  true,
						"priority": 100,
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot marshall app config: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("cannot write app config into a file: %w", err)
	}

	return nil
}

func writePolicyConfig(path string) error {
	cfg := map[string]any{
		"enabled":    false,
		"priority":   10,
		"whitelist":  "./data/whitelist.txt",
		"blacklist":  "./data/blacklist.txt",
		"action_dir": "actions",
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot marshall policy config: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("cannot write policy config into a file: %w", err)
	}

	return nil
}
