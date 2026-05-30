package plugin

import (
	"fmt"
	"sync"
)

type PluginFactory func() Plugin

var (
	registry = map[string]PluginFactory{}
	mu       sync.RWMutex
)

func Register(name string, factory PluginFactory) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := registry[name]; exists {
		return fmt.Errorf("plugin already registered: %s", name)
	}

	registry[name] = factory

	return nil
}

func resetRegistry() {
	mu.Lock()
	defer mu.Unlock()

	registry = map[string]PluginFactory{}
}
