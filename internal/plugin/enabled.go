package plugin

const DefaultEnabled = true

type EnabledPlugin interface {
	Enabled() *bool
}

func getEnabled(p Plugin) bool {
	var enabled *bool
	if e, ok := p.(EnabledPlugin); ok {
		enabled = e.Enabled()
	}

	if enabled != nil {
		return *enabled
	}

	return DefaultEnabled
}
