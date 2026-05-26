package config

import "errors"

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
	Session     string
}

func ValidateWorkspaceFlags(wf WorkspaceFlags) error {
	if wf.Session != "" && wf.Workspace == "" {
		return errors.New("Passed session ID requires workspace name")
	}

	return nil
}
