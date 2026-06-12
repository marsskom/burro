package lua

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"

	"gitlab.com/marsskom/burro/internal/lua/pkg"
	"gitlab.com/marsskom/burro/internal/lua/pluginwrapper"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

var (
	ErrLuaScriptDirectoryNotFound    = errors.New("lua script directory not found")
	ErrLuaScriptDirectoryNotReadable = errors.New("lua script directory cannot be read")
	ErrLuaScriptFileNotReadable      = errors.New("lua script file cannot be read")
	ErrLuaScriptCallNotFound         = errors.New("lua script's call not found")

	ErrLuaScriptExecute = errors.New("error on lua script execution")
)

var eventScriptMap = map[string]string{
	"flush":           "flush",
	"connect":         "connect",
	"before_request":  "before_request",
	"after_request":   "after_request",
	"before_response": "before_response",
	"after_response":  "after_response",
	"error":           "error",
	"close":           "close",

	"ws_open":  "ws_open",
	"ws_msg":   "ws_msg",
	"ws_close": "ws_close",
}

type PluginWrapper struct {
	runtime pluginapi.Runtime
	dir     string

	lStates map[string]*lua.LState
	lCalls  map[string]map[string]*lua.LFunction
}

func NewPluginWrapper(rt pluginapi.Runtime, dir string) *PluginWrapper {
	return &PluginWrapper{
		runtime: rt,
		dir:     dir,
		lStates: make(map[string]*lua.LState),
		lCalls:  make(map[string]map[string]*lua.LFunction),
	}
}

func (w *PluginWrapper) InitState(script string) error {
	if _, ok := w.lStates[script]; ok {
		return fmt.Errorf("lua plugin wrapper cannot init state for existing on '%s'", script)
	}

	L := newSandboxState()

	// Packages.
	pkg.RegisterTime(L)
	pkg.RegisterBase64(L)
	pkg.RegisterJSON(L)

	// Runtime.
	pluginwrapper.RegisterLogger(L, w.runtime.Log())
	pluginwrapper.RegisterKeyValueStore(L, w.runtime.KV())
	pluginwrapper.RegisterEventBus(L, w.runtime.Events())
	pluginwrapper.RegisterArtifactsStore(L, w.runtime.Artifacts())
	pluginwrapper.RegisterDataStore(L, w.runtime.Data())

	w.lStates[script] = L

	return nil
}

func (w *PluginWrapper) InitCalls(script string) error {
	if _, ok := w.lStates[script]; !ok {
		return fmt.Errorf("lua plugin wrapper: you need init state for the script '%s' first to init calls", script)
	}

	path := filepath.Join(w.dir, script)
	if !w.runtime.Data().Exists(path) {
		return fmt.Errorf("lua plugin wrapper cannot find script directory '%s': %w", path, ErrLuaScriptDirectoryNotFound)
	}

	list, err := w.runtime.Data().List(path, []string{".lua"})
	if err != nil {
		return fmt.Errorf("lua plugin wrapper cannot read script directory '%s': %w", path, ErrLuaScriptDirectoryNotReadable)
	}

	L := w.lStates[script]
	calls := make(map[string]*lua.LFunction)

	for _, name := range list {
		f, err := w.runtime.Data().Read(name)
		if err != nil {
			return fmt.Errorf("lua plugin wrapper cannot read script file '%s': %w", name, ErrLuaScriptFileNotReadable)
		}

		b, err := io.ReadAll(f)
		if err != nil {
			f.Close()

			return fmt.Errorf("lua plugin wrapper cannot read script file '%s': %w", name, ErrLuaScriptFileNotReadable)
		}
		f.Close()

		fn, err := L.LoadString(string(b))
		if err != nil {
			return fmt.Errorf("lua plugin wrapper cannot load file '%s'", name)
		}

		callName := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))

		w.runtime.Log().Debug("lua plugin wrapper: registers call for a script", "call", callName, "script", script)

		calls[callName] = fn
	}

	w.lCalls[script] = calls

	return nil
}

func (w *PluginWrapper) Shutdown() error {
	w.lCalls = make(map[string]map[string]*lua.LFunction)

	for _, L := range w.lStates {
		L.Close()
	}

	return nil
}
