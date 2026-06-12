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

func isPluginEnabled(cfg any) bool {
	if cfg == nil {
		return DefaultEnabled
	}

	// If config is a map.
	if m, ok := cfg.(map[string]any); ok {
		if v, ok := m["enabled"]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}

	return DefaultEnabled
}
