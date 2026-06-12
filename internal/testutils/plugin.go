package testutils

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"gitlab.com/marsskom/burro/internal/pluginapi"
	rt "gitlab.com/marsskom/burro/internal/runtime"
)

type TestRuntime struct {
	artifacts pluginapi.ArtifactStore
	data      pluginapi.DataStore
	kv        pluginapi.KeyValueStore
	eventBus  pluginapi.EventBus
	log       pluginapi.Logger
}

func NewRuntime(
	a pluginapi.ArtifactStore,
	d pluginapi.DataStore,
	kv pluginapi.KeyValueStore,
	e pluginapi.EventBus,
	l pluginapi.Logger,
) pluginapi.Runtime {
	return &TestRuntime{
		artifacts: a,
		data:      d,
		kv:        kv,
		eventBus:  e,
		log:       l,
	}
}

func NewForPlugin(home string) pluginapi.Runtime {
	return &TestRuntime{
		rt.NewFileArtifactStore(filepath.Join(home)),
		rt.NewPluginDataStore(filepath.Join(home)),
		rt.NewKeyValue(),
		rt.NewEventBus(),
		NewMemoryLogger(),
	}
}

func (r *TestRuntime) Artifacts() pluginapi.ArtifactStore { return r.artifacts }
func (r *TestRuntime) Data() pluginapi.DataStore          { return r.data }
func (r *TestRuntime) KV() pluginapi.KeyValueStore        { return r.kv }
func (r *TestRuntime) Events() pluginapi.EventBus         { return r.eventBus }
func (r *TestRuntime) Log() pluginapi.Logger              { return r.log }

type MemoryLogger struct {
	Messages map[string][]string

	mu sync.Mutex
}

func NewMemoryLogger() *MemoryLogger {
	return &MemoryLogger{
		Messages: map[string][]string{
			"trace": make([]string, 0),
			"debug": make([]string, 0),
			"info":  make([]string, 0),
			"warn":  make([]string, 0),
			"error": make([]string, 0),
			"audit": make([]string, 0),
		},
	}
}

func (l *MemoryLogger) Enabled(level slog.Level) bool {
	return true
}

func (l *MemoryLogger) Trace(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Messages["trace"] = append(l.Messages["trace"], fmt.Sprintf("%s ; args: %+v", msg, args))
}

func (l *MemoryLogger) Debug(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Messages["debug"] = append(l.Messages["debug"], fmt.Sprintf("%s ; args: %+v", msg, args))
}
func (l *MemoryLogger) Info(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Messages["info"] = append(l.Messages["info"], fmt.Sprintf("%s ; args: %+v", msg, args))
}
func (l *MemoryLogger) Warn(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Messages["warn"] = append(l.Messages["warn"], fmt.Sprintf("%s ; args: %+v", msg, args))
}
func (l *MemoryLogger) Error(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Messages["error"] = append(l.Messages["error"], fmt.Sprintf("%s ; args: %+v", msg, args))
}
func (l *MemoryLogger) Audit(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Messages["audit"] = append(l.Messages["audit"], fmt.Sprintf("%s ; args: %+v", msg, args))
}
