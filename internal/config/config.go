package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var userHomeDir = os.UserHomeDir

type CoreConfig struct {
	LogLevel string `yaml:"log_level"`
}

type ProxyConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type Config struct {
	Core    CoreConfig     `yaml:"core"`
	Proxy   ProxyConfig    `yaml:"proxy"`
	Plugins map[string]any `yaml:"plugins"`
}

func LoadWithFlags(configPath string, proxyFlags ProxyFlags) (*Config, error) {
	cfg, err := Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("Config: cannot load config: %w", err)
	}

	cfg.Proxy = MergeProxy(cfg.Proxy, proxyFlags)

	return cfg, nil
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Config: caanot read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Config: cannot unmarshall config file: %w", err)
	}

	return &cfg, nil
}

func ResolvePath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	if env := os.Getenv("BURRO_CONFIG"); env != "" {
		return env, nil
	}

	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve config path error on get user home dir: %w", err)
	}

	defaultPath := filepath.Join(home, ".burro", "config.yml")

	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}

	if _, err := os.Stat("./config.yml"); err == nil {
		return "./config.yml", nil
	}

	return "", errors.New("config not found")
}
