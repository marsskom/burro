package config

type ProxyFlags struct {
	ZeroCfg    bool
	Listen     string
	GRPCListen string
	WorkDir    string
	Workspace  string
	TLSCert    string
	TLSKey     string
	CACert     string
	CAKey      string
}
