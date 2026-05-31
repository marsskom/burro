package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

func ResolveHome(explicit string) string {
	if explicit != "" {
		return explicit
	}

	if env := os.Getenv("BURRO_HOME"); env != "" {
		return env
	}

	return "./runtime"
}

func (p *Paths) GetConfigPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	if env := os.Getenv("BURRO_CONFIG"); env != "" {
		return env, nil
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
	Plugins map[string]any `yaml:"plugins"`
}

type CoreConfig struct {
	LogLevel string            `yaml:"log_level"`
	Plugins  CorePluginsConfig `yaml:"plugins"`
}

type ProxyConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type CorePluginsConfig struct {
	Dir    string `yaml:"dir"`
	Config string `yaml:"config"`
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
		return nil, fmt.Errorf("Config: cannot read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Config: cannot unmarshall config file: %w", err)
	}

	return &cfg, nil
}
