package plugin

import (
	"sync"
	"testing"

	"gitlab.com/marsskom/burro/internal/pluginapi"
)

type mockPlugin struct {
	initCalled bool
	nameCalled bool
}

func (m *mockPlugin) Shutdown() error { return nil }

func (mp *mockPlugin) Init(rt pluginapi.Runtime, cfg any) error {
	mp.initCalled = true

	return nil
}

func (mp *mockPlugin) Name() string {
	mp.nameCalled = true

	return "mockPlugin"
}

func TestRegister_Success(t *testing.T) {
	resetRegistry()

	err := Register("test", func() Plugin {
		return &mockPlugin{}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	resetRegistry()

	_ = Register("test", func() Plugin {
		return &mockPlugin{}
	})

	err := Register("test", func() Plugin {
		return &mockPlugin{}
	})

	if err == nil {
		t.Fatal("expected duplicate registration error, got nil")
	}
}

func TestRegister_ErrorMessage(t *testing.T) {
	resetRegistry()

	_ = Register("test", func() Plugin {
		return &mockPlugin{}
	})

	err := Register("test", func() Plugin {
		return &mockPlugin{}
	})

	expected := "plugin already registered: test"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestRegister_ConcurrentSafety(t *testing.T) {
	resetRegistry()

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			err := Register("test", func() Plugin {
				return &mockPlugin{}
			})

			errors <- err
		}()
	}

	wg.Wait()
	close(errors)

	var success, failures int

	for err := range errors {
		if err != nil {
			failures++
		} else {
			success++
		}
	}

	if success != 1 {
		t.Fatalf("expected exactly 1 success, got %d", success)
	}

	if failures != goroutines-1 {
		t.Fatalf("expected %d failures, got %d", goroutines-1, failures)
	}
}
