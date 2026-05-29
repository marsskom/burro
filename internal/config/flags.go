package config

type ProxyFlags struct {
	Port int
}

func MergeProxy(cfg ProxyConfig, flags ProxyFlags) ProxyConfig {
	if flags.Port != 0 {
		cfg.Port = flags.Port
	}

	return cfg
}

type WorkspaceFlags struct {
	Interactive bool
	Workspace   string
}
