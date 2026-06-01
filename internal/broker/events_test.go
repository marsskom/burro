package broker

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestHub_SubscribeAndPublish(t *testing.T) {
	h := NewHub()

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	event := Event{
		Type: EventRequest,
		ID:   "1",
	}

	h.Publish(event)

	select {
	case got := <-ch:
		if got.ID != "1" {
			t.Fatalf("expected ID=1, got %s", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestHub_MultipleSubscribers(t *testing.T) {
	h := NewHub()

	ch1 := h.Subscribe()
	ch2 := h.Subscribe()

	defer h.Unsubscribe(ch1)
	defer h.Unsubscribe(ch2)

	e := Event{ID: "x"}

	h.Publish(e)

	for _, ch := range []chan Event{ch1, ch2} {
		select {
		case got := <-ch:
			if got.ID != "x" {
				t.Fatalf("wrong event: %s", got.ID)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}
}

func TestHub_Unsubscribe_ClosesChannel(t *testing.T) {
	h := NewHub()

	ch := h.Subscribe()

	h.Unsubscribe(ch)

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed")
	}
}

func TestHub_Publish_NoSubscribers(t *testing.T) {
	h := NewHub()

	// Must not panic or deadlock.
	h.Publish(Event{ID: "x"})
}

func TestHub_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	h := NewHub()

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			ch := h.Subscribe()

			time.Sleep(time.Millisecond)

			h.Unsubscribe(ch)
		}()
	}

	wg.Wait()
}

func TestHub_SlowSubscriber_DoesNotBlockPublish(t *testing.T) {
	h := NewHub()

	slow := h.Subscribe()
	defer h.Unsubscribe(slow)

	// Fills buffer to simulate slow consumer.
	for i := 0; i < 100; i++ {
		select {
		case slow <- Event{}:
		default:
		}
	}

	done := make(chan struct{})

	go func() {
		h.Publish(Event{ID: "x"})
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(time.Second):
		t.Fatal("publish blocked")
	}
}

func TestHub_ConcurrentPublish(t *testing.T) {
	h := NewHub()

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	var wg sync.WaitGroup

	for i := range 50 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			h.Publish(Event{ID: fmt.Sprintf("%d", i)})
		}(i)
	}

	wg.Wait()

	received := 0

	drain := true
	for drain {
		select {
		case <-ch:
			received++
		default:
			drain = false
		}
	}

	if received == 0 {
		t.Fatal("expected some events")
	}
}
