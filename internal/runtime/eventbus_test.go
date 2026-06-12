package runtime

import (
	"sync"
	"testing"
	"time"

	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func TestEventBus_Emit(t *testing.T) {
	bus := NewEventBus()

	done := make(chan pluginapi.Event, 1)

	bus.On("test", func(e pluginapi.Event) {
		done <- e
	})

	err := bus.Emit(pluginapi.Event{
		Name: "test",
		Data: "hello",
		From: "me",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case e := <-done:
		if e.Data != "hello" {
			t.Fatalf("expected hello, got %v", e.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_Emit_DifferentEventName(t *testing.T) {
	bus := NewEventBus()

	called := make(chan struct{}, 1)

	bus.On("foo", func(e pluginapi.Event) {
		called <- struct{}{}
	})

	_ = bus.Emit(pluginapi.Event{
		Name: "bar",
	})

	select {
	case <-called:
		t.Fatal("handler should not be called")
	case <-time.After(100 * time.Millisecond):
		// OK
	}
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	bus := NewEventBus()

	var wg sync.WaitGroup
	wg.Add(2)

	bus.On("test", func(e pluginapi.Event) {
		wg.Done()
	})

	bus.On("test", func(e pluginapi.Event) {
		wg.Done()
	})

	_ = bus.Emit(pluginapi.Event{Name: "test"})

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handlers were not called")
	}
}

func TestEventBus_MultipleEventTypes(t *testing.T) {
	bus := NewEventBus()

	a := make(chan struct{}, 1)
	b := make(chan struct{}, 1)

	bus.On("a", func(e pluginapi.Event) {
		a <- struct{}{}
	})

	bus.On("b", func(e pluginapi.Event) {
		b <- struct{}{}
	})

	_ = bus.Emit(pluginapi.Event{Name: "a"})

	select {
	case <-a:
	case <-time.After(time.Second):
		t.Fatal("event a not received")
	}

	select {
	case <-b:
		t.Fatal("event b should not be received")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestEventBus_ConcurrentEmit(t *testing.T) {
	bus := NewEventBus()

	var wg sync.WaitGroup

	bus.On("test", func(e pluginapi.Event) {
		wg.Done()
	})

	const events = 100

	wg.Add(events)

	for range events {
		go func() {
			_ = bus.Emit(pluginapi.Event{
				Name: "test",
			})
		}()
	}

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestEventBus_Off(t *testing.T) {
	bus := NewEventBus()

	called := make(chan struct{}, 1)

	off := bus.On("test", func(e pluginapi.Event) {
		called <- struct{}{}
	})

	off() // unsubscribe

	_ = bus.Emit(pluginapi.Event{Name: "test"})

	select {
	case <-called:
		t.Fatal("handler should not be called after Off")
	case <-time.After(100 * time.Millisecond):
		// OK
	}
}
