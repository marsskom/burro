package plugin

var registry = map[string]PluginFactory{}

func Register(name string, factory PluginFactory) {
	registry[name] = factory
}
