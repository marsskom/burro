package config

type ProxyFlags struct {
	ZeroCfg      bool
	Listen       string
	GRPCListen   string
	GRPCDisabled bool
	GRPCDebug    bool
	WorkDir      string
	Workspace    string
	TLSCert      string
	TLSKey       string
	TLSInsecure  bool
	CADisabled   bool
	CACert       string
	CAKey        string
}
