package pluginwrapper

import (
	"log/slog"
	"testing"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

type logEntry struct {
	level string
	msg   string
	args  []any
}

type mockLogger struct {
	entries []logEntry
}

func (m *mockLogger) log(level, msg string, args []any) {
	m.entries = append(m.entries, logEntry{level: level, msg: msg, args: args})
}

func (m *mockLogger) Enabled(level slog.Level) bool { return true }
func (m *mockLogger) Trace(msg string, args ...any) { m.log("trace", msg, args) }
func (m *mockLogger) Debug(msg string, args ...any) { m.log("debug", msg, args) }
func (m *mockLogger) Info(msg string, args ...any)  { m.log("info", msg, args) }
func (m *mockLogger) Warn(msg string, args ...any)  { m.log("warn", msg, args) }
func (m *mockLogger) Error(msg string, args ...any) { m.log("error", msg, args) }
func (m *mockLogger) Audit(msg string, args ...any) { m.log("audit", msg, args) }

func newLogState(t *testing.T, log pluginapi.Logger) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })

	RegisterLogger(L, log)

	return L
}

func assertEntry(t *testing.T, got logEntry, wantLevel, wantMsg string) {
	t.Helper()
	if got.level != wantLevel {
		t.Errorf("level: want %s got %s", wantLevel, got.level)
	}
	if got.msg != wantMsg {
		t.Errorf("msg: want %q got %q", wantMsg, got.msg)
	}
}

// Each level is routed correctly.

func TestLogger_Trace(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`log.trace("hello trace")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if len(log.entries) != 1 {
		t.Fatalf("expected 1 entry got %d", len(log.entries))
	}
	assertEntry(t, log.entries[0], "trace", "hello trace")
}

func TestLogger_Debug(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`log.debug("hello debug")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	assertEntry(t, log.entries[0], "debug", "hello debug")
}

func TestLogger_Info(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`log.info("hello info")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	assertEntry(t, log.entries[0], "info", "hello info")
}

func TestLogger_Warn(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`log.warn("hello warn")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	assertEntry(t, log.entries[0], "warn", "hello warn")
}

func TestLogger_Error(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`log.error("hello error")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	assertEntry(t, log.entries[0], "error", "hello error")
}

func TestLogger_Audit(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`log.audit("hello audit")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	assertEntry(t, log.entries[0], "audit", "hello audit")
}

// Attrs table are passed as key-value pairs.

func TestLogger_WithAttrs(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`log.info("msg", { url = "http://example.com", status = "200" })`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	args := log.entries[0].args
	if len(args) != 4 {
		t.Fatalf("args: want 4 got %d: %v", len(args), args)
	}
	// Args are k,v,k,v — finds url key.
	found := false
	for i := 0; i < len(args)-1; i += 2 {
		if args[i] == "url" && args[i+1] == "http://example.com" {
			found = true
		}
	}
	if !found {
		t.Errorf("url attr not found in args: %v", args)
	}
}

func TestLogger_WithAttrs_Nil(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`log.info("no attrs")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if len(log.entries[0].args) != 0 {
		t.Errorf("args: want empty got %v", log.entries[0].args)
	}
}

// No return values.

func TestLogger_ReturnsNothing(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	// If log.info returned something this would panic or error.
	if err := L.DoString(`local x = log.info("test"); assert(x == nil)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
}

// Multiple calls in one script.

func TestLogger_MultipleCalls(t *testing.T) {
	log := &mockLogger{}
	L := newLogState(t, log)
	if err := L.DoString(`
		log.debug("first")
		log.info("second")
		log.warn("third")
		log.error("fourth")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if len(log.entries) != 4 {
		t.Fatalf("expected 4 entries got %d", len(log.entries))
	}
	assertEntry(t, log.entries[0], "debug", "first")
	assertEntry(t, log.entries[1], "info", "second")
	assertEntry(t, log.entries[2], "warn", "third")
	assertEntry(t, log.entries[3], "error", "fourth")
}
