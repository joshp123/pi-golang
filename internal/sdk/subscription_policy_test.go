package sdk

import (
	"errors"
	"testing"
	"time"
)

func TestSubscriptionPolicyDropEmitsDiagnostic(t *testing.T) {
	hub := newEventHub()
	events, cancel, err := hub.Subscribe(toStreamPolicy(SubscriptionPolicy{Buffer: 1, Mode: SubscriptionModeDrop, EmitDropEvent: true}))
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	defer cancel()

	for index := 0; index < 8; index++ {
		hub.Publish(Event{Type: EventTypeMessageUpdate, Raw: []byte(`{"type":"message_update"}`)})
	}

	deadline := time.After(1 * time.Second)
	for {
		select {
		case event := <-events:
			if event.Type == EventTypeSubscriptionDrop {
				return
			}
		case <-deadline:
			t.Fatal("expected subscription_drop event")
		}
	}
}

func TestSubscriptionPolicyRingRetainsNewest(t *testing.T) {
	hub := newEventHub()
	events, cancel, err := hub.Subscribe(toStreamPolicy(SubscriptionPolicy{Buffer: 1, Mode: SubscriptionModeRing}))
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	defer cancel()

	hub.Publish(Event{Type: "e1", Raw: []byte(`{"type":"e1"}`)})
	hub.Publish(Event{Type: "e2", Raw: []byte(`{"type":"e2"}`)})
	hub.Publish(Event{Type: "e3", Raw: []byte(`{"type":"e3"}`)})

	deadline := time.After(1 * time.Second)
	seenNewest := false
	for i := 0; i < 3; i++ {
		select {
		case event := <-events:
			if event.Type == "e3" {
				seenNewest = true
			}
		case <-deadline:
			if !seenNewest {
				t.Fatal("expected ring mode to retain newest event")
			}
			return
		}
	}
	if !seenNewest {
		t.Fatal("expected ring mode to retain newest event")
	}
}

func TestSubscriptionPolicyBlockPreservesAll(t *testing.T) {
	hub := newEventHub()
	events, cancel, err := hub.Subscribe(toStreamPolicy(SubscriptionPolicy{Buffer: 1, Mode: SubscriptionModeBlock}))
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	defer cancel()

	hub.Publish(Event{Type: "e1", Raw: []byte(`{"type":"e1"}`)})

	done := make(chan struct{})
	go func() {
		hub.Publish(Event{Type: "e2", Raw: []byte(`{"type":"e2"}`)})
		hub.Publish(Event{Type: "e3", Raw: []byte(`{"type":"e3"}`)})
		close(done)
	}()

	if got := readQueuedEvent(t, events).Type; got != "e1" {
		t.Fatalf("expected e1, got %s", got)
	}
	if got := readQueuedEvent(t, events).Type; got != "e2" {
		t.Fatalf("expected e2, got %s", got)
	}
	if got := readQueuedEvent(t, events).Type; got != "e3" {
		t.Fatalf("expected e3, got %s", got)
	}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("block mode did not preserve all events")
	}
}

func TestSubscriptionPolicyRejectsInvalidConfig(t *testing.T) {
	err := validateSubscriptionPolicy(SubscriptionPolicy{Buffer: 0, Mode: SubscriptionModeDrop})
	if !errors.Is(err, ErrInvalidSubscriptionPolicy) {
		t.Fatalf("expected ErrInvalidSubscriptionPolicy, got %v", err)
	}

	err = validateSubscriptionPolicy(SubscriptionPolicy{Buffer: 4, Mode: SubscriptionMode("invalid")})
	if !errors.Is(err, ErrInvalidSubscriptionPolicy) {
		t.Fatalf("expected ErrInvalidSubscriptionPolicy, got %v", err)
	}
}

func readQueuedEvent(t *testing.T, ch <-chan Event) Event {
	t.Helper()
	select {
	case event := <-ch:
		return event
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for queued event")
		return Event{}
	}
}
