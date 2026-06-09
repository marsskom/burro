package actions

type ActionFile struct {
	Actions []ActionRule `yaml:"actions"`
}

type OnMatchType string

const (
	OnMatchStop     OnMatchType = "stop"
	OnMatchContinue OnMatchType = "continue"
)

type ActionRule struct {
	ID       string      `yaml:"id"`
	Priority int         `yaml:"priority"`
	Match    Match       `yaml:"match"`
	Action   []Action    `yaml:"action"`
	OnMatch  OnMatchType `yaml:"on_match"`
}

type Match struct {
	All []Match `yaml:"all,omitempty"`
	Any []Match `yaml:"any,omitempty"`
	Not *Match  `yaml:"not,omitempty"`

	Method string `yaml:"method,omitempty"`
	Domain string `yaml:"domain,omitempty"`
	Path   string `yaml:"path,omitempty"`
	IP     string `yaml:"ip,omitempty"`

	Headers map[string]string `yaml:"headers,omitempty"`
}

type Operation string

const (
	OpDeny         Operation = "deny"
	OpAllow        Operation = "allow"
	OpSetHeader    Operation = "set_header"
	OpRemoveHeader Operation = "remove_header"
	OpRedactBody   Operation = "redact_body"
	OpLog          Operation = "log"
)

type Action struct {
	Operation Operation `yaml:"op"`
	Args      any       `yaml:"args"`
}

type SetHeaderArgs struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type RemoveHeaderArgs struct {
	Names []string `yaml:"names"`
}

type RedactBodyArgs struct {
	Fields []string `yaml:"fields"`
}

type LogArgs struct {
	Level   string `yaml:"level"`
	Message string `yaml:"message"`
}
