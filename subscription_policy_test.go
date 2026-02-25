package pi

import (
	"errors"
	"testing"
	"time"
)

func TestSubscriptionPolicyDropEmitsDiagnostic(t *testing.T) {
	sub, err := newSubscription(SubscriptionPolicy{Buffer: 1, Mode: SubscriptionModeDrop, EmitDropEvent: true})
	if err != nil {
		t.Fatalf("newSubscription returned error: %v", err)
	}
	subscribers := map[*subscription]struct{}{sub: {}}

	publishToSubscribers(subscribers, Event{Type: EventTypeMessageUpdate, Raw: []byte(`{"type":"message_update"}`)})
	publishToSubscribers(subscribers, Event{Type: EventTypeMessageUpdate, Raw: []byte(`{"type":"message_update"}`)})

	event := readQueuedEvent(t, sub.in)
	if event.Type != EventTypeSubscriptionDrop {
		t.Fatalf("expected %s, got %s", EventTypeSubscriptionDrop, event.Type)
	}
}

func TestSubscriptionPolicyRingRetainsNewest(t *testing.T) {
	sub, err := newSubscription(SubscriptionPolicy{Buffer: 1, Mode: SubscriptionModeRing})
	if err != nil {
		t.Fatalf("newSubscription returned error: %v", err)
	}
	subscribers := map[*subscription]struct{}{sub: {}}

	publishToSubscribers(subscribers, Event{Type: "e1", Raw: []byte(`{"type":"e1"}`)})
	publishToSubscribers(subscribers, Event{Type: "e2", Raw: []byte(`{"type":"e2"}`)})
	publishToSubscribers(subscribers, Event{Type: "e3", Raw: []byte(`{"type":"e3"}`)})

	event := readQueuedEvent(t, sub.in)
	if event.Type != "e3" {
		t.Fatalf("expected newest event e3, got %s", event.Type)
	}
}

func TestSubscriptionPolicyBlockPreservesAll(t *testing.T) {
	sub, err := newSubscription(SubscriptionPolicy{Buffer: 1, Mode: SubscriptionModeBlock})
	if err != nil {
		t.Fatalf("newSubscription returned error: %v", err)
	}
	subscribers := map[*subscription]struct{}{sub: {}}

	publishToSubscribers(subscribers, Event{Type: "e1", Raw: []byte(`{"type":"e1"}`)})

	done := make(chan struct{})
	go func() {
		publishToSubscribers(subscribers, Event{Type: "e2", Raw: []byte(`{"type":"e2"}`)})
		publishToSubscribers(subscribers, Event{Type: "e3", Raw: []byte(`{"type":"e3"}`)})
		close(done)
	}()

	if got := readQueuedEvent(t, sub.in).Type; got != "e1" {
		t.Fatalf("expected e1, got %s", got)
	}
	if got := readQueuedEvent(t, sub.in).Type; got != "e2" {
		t.Fatalf("expected e2, got %s", got)
	}
	if got := readQueuedEvent(t, sub.in).Type; got != "e3" {
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
