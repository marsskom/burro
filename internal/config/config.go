package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CoreConfig struct {
	logLevel string `yaml:"log_level"`
}

type ProxyConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type Config struct {
	Core  CoreConfig  `yaml:"core"`
	Proxy ProxyConfig `yaml:"proxy"`
}

func LoadWithFlags(configPath string, proxyFlags ProxyFlags) (*Config, error) {
	cfg, err := Load(configPath)
	if err != nil {
		return nil, err
	}

	cfg.Proxy = MergeProxy(cfg.Proxy, proxyFlags)

	return cfg, nil
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
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

	home, _ := os.UserHomeDir()
	defaultPath := filepath.Join(home, ".burro", "config.yml")

	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}

	if _, err := os.Stat("./config.yml"); err == nil {
		return "./config.yml", nil
	}

	return "", errors.New("config not found")
}
