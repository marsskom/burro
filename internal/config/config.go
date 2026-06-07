package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"gitlab.com/marsskom/burro/internal/logger"
	"gopkg.in/yaml.v3"
)

type Paths struct {
	Home string
}

func NewPaths(home string) *Paths {
	return &Paths{
		Home: home,
	}
}

func ResolveWorkdir(explicit string) string {
	if explicit != "" {
		return explicit
	}

	if env := os.Getenv("BURRO_WORKDIR"); env != "" {
		return env
	}

	return "./runtime"
}

func (p *Paths) GetConfigPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	path := filepath.Join(p.Home, "config.yml")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	return "", errors.New("config not found")
}

type Config struct {
	Core    CoreConfig     `yaml:"core"`
	Proxy   ProxyConfig    `yaml:"proxy"`
	GRPC    GRPCConfig     `yaml:"grpc"`
	TLS     TLSConfig      `yaml:"tls"`
	Plugins map[string]any `yaml:"plugins"`
}

type CoreConfig struct {
	LogLevel string            `yaml:"log_level"`
	Plugins  CorePluginsConfig `yaml:"plugins"`
}

type CorePluginsConfig struct {
	Dir    string `yaml:"dir"`
	Config string `yaml:"config"`
}

type ProxyConfig struct {
	ZeroConfigurationMode bool
	Listen                string `yaml:"listen"`
}

type GRPCConfig struct {
	Enabled bool   `yaml:"enabled"`
	Debug   bool   `yaml:"debug"`
	Listen  string `yaml:"listen"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Cert     string `yaml:"cert"`
	Key      string `yaml:"key"`
	Insecure bool   `yaml:"insecure"`
}

func NewZeroCfg(flags ProxyFlags) (*Config, error) {
	return mergeProxyFlags(&Config{
		Plugins: map[string]any{
			"logger": make(map[string]any),
		},
	}, flags), nil
}

func LoadWithFlags(configPath string, flags ProxyFlags) (*Config, error) {
	cfg, err := Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot load config: %w", err)
	}

	cfg = mergeProxyFlags(cfg, flags)

	return cfg, nil
}

func mergeProxyFlags(cfg *Config, flags ProxyFlags) *Config {
	cfg.Proxy.ZeroConfigurationMode = flags.ZeroCfg

	if flags.TLSCert != "" {
		cfg.TLS.Enabled = true
		cfg.TLS.Cert = flags.TLSCert
		cfg.TLS.Key = flags.TLSKey
	}

	if flags.TLSInsecure {
		cfg.TLS.Insecure = true
	}

	if flags.Listen != "" {
		host, port, err := net.SplitHostPort(flags.Listen)
		if err != nil {
			logger.Warn("error parsing listen argument, run on config settings", "error", err)
		} else {
			cfg.Proxy.Listen = fmt.Sprintf("%s:%s", host, port)
		}
	}

	if flags.GRPCListen != "" {
		host, port, err := net.SplitHostPort(flags.GRPCListen)
		if err != nil {
			logger.Warn("error parsing gRPC listen argument, run on config settings", "error", err)
		} else {
			cfg.GRPC.Enabled = true
			cfg.GRPC.Listen = fmt.Sprintf("%s:%s", host, port)
		}
	}

	if flags.GRPCDisabled {
		cfg.GRPC.Enabled = false
	}

	if flags.GRPCDebug {
		cfg.GRPC.Debug = true
	}

	return cfg
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshall config file: %w", err)
	}

	return &cfg, nil
}
