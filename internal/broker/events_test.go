package broker

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestHub_SubscribeAndPublish(t *testing.T) {
	h := NewHub()

	sub := h.Subscribe([]TransportType{}, []EventType{})
	defer h.Unsubscribe(sub)

	event := Event{
		Type: EventBeforeRequestSend,
		ID:   "1",
	}

	h.Publish(event)

	select {
	case got := <-sub.Ch:
		if got.ID != "1" {
			t.Fatalf("expected ID=1, got %s", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestHub_MultipleSubscribers(t *testing.T) {
	h := NewHub()

	sub1 := h.Subscribe([]TransportType{}, []EventType{})
	sub2 := h.Subscribe([]TransportType{}, []EventType{})

	defer h.Unsubscribe(sub1)
	defer h.Unsubscribe(sub2)

	e := Event{ID: "x"}

	h.Publish(e)

	for _, ch := range []chan Event{sub1.Ch, sub2.Ch} {
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

func expect(t *testing.T, name string, ch chan Event, shouldReceive bool) {
	t.Helper()

	select {
	case ev := <-ch:
		if !shouldReceive {
			t.Fatalf("%s: unexpected event %s", name, ev.ID)
		}
	case <-time.After(200 * time.Millisecond):
		if shouldReceive {
			t.Fatalf("%s: expected event but got timeout", name)
		}
	}
}

func TestHub_FilteringMatrix(t *testing.T) {
	h := NewHub()

	sub1 := h.Subscribe([]TransportType{TransportHTTP}, []EventType{})
	sub2 := h.Subscribe([]TransportType{TransportWS}, []EventType{})
	sub3 := h.Subscribe([]TransportType{}, []EventType{EventConnect, EventWSConnect})

	defer h.Unsubscribe(sub1)
	defer h.Unsubscribe(sub2)
	defer h.Unsubscribe(sub3)

	httpConnect := Event{
		ID: "http-connect", Type: EventConnect, TransportType: TransportHTTP,
	}
	wsConnect := Event{
		ID: "ws-connect", Type: EventWSConnect, TransportType: TransportWS,
	}
	httpOther := Event{
		ID: "http-other", Type: EventError, TransportType: TransportHTTP,
	}

	// Fires events one by one (important for deterministic reads).
	h.Publish(httpConnect)

	expect(t, "sub1 httpConnect", sub1.Ch, true)
	expect(t, "sub2 httpConnect", sub2.Ch, false)
	expect(t, "sub3 httpConnect", sub3.Ch, true)

	h.Publish(wsConnect)

	expect(t, "sub1 wsConnect", sub1.Ch, false)
	expect(t, "sub2 wsConnect", sub2.Ch, true)
	expect(t, "sub3 wsConnect", sub3.Ch, true)

	h.Publish(httpOther)

	expect(t, "sub1 httpOther", sub1.Ch, true)
	expect(t, "sub2 httpOther", sub2.Ch, false)
	expect(t, "sub3 httpOther", sub3.Ch, false)
}

func TestHub_Unsubscribe_ClosesChannel(t *testing.T) {
	h := NewHub()

	sub := h.Subscribe([]TransportType{}, []EventType{})

	h.Unsubscribe(sub)

	_, ok := <-sub.Ch
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

			sub := h.Subscribe([]TransportType{}, []EventType{})

			time.Sleep(time.Millisecond)

			h.Unsubscribe(sub)
		}()
	}

	wg.Wait()
}

func TestHub_SlowSubscriber_DoesNotBlockPublish(t *testing.T) {
	h := NewHub()

	slow := h.Subscribe([]TransportType{}, []EventType{})
	defer h.Unsubscribe(slow)

	// Fills buffer to simulate slow consumer.
	for range 100 {
		select {
		case slow.Ch <- Event{}:
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

	sub := h.Subscribe([]TransportType{}, []EventType{})
	defer h.Unsubscribe(sub)

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
		case <-sub.Ch:
			received++
		default:
			drain = false
		}
	}

	if received == 0 {
		t.Fatal("expected some events")
	}
}
