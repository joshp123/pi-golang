package sdk

import (
	"testing"
	"time"
)

func TestEventHubCloseUnblocksBlockedPublisher(t *testing.T) {
	hub := newEventHub()
	events, _, err := hub.Subscribe(toStreamPolicy(SubscriptionPolicy{Buffer: 1, Mode: SubscriptionModeBlock}))
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	_ = events

	hub.Publish(Event{Type: "e1", Raw: []byte(`{"type":"e1"}`)})

	done := make(chan struct{})
	go func() {
		hub.Publish(Event{Type: "e2", Raw: []byte(`{"type":"e2"}`)})
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	hub.Close()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("publish remained blocked after hub close")
	}
}

func TestEventHubProcessDiedPublishesAndCloses(t *testing.T) {
	hub := newEventHub()
	events, _, err := hub.Subscribe(toStreamPolicy(SubscriptionPolicy{Buffer: 2, Mode: SubscriptionModeRing}))
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	hub.ProcessDied(Event{Type: EventTypeProcessDied, Raw: []byte(`{"type":"process_died"}`)})

	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("expected process_died event before close")
		}
		if event.Type != EventTypeProcessDied {
			t.Fatalf("unexpected event type: %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for process_died event")
	}

	select {
	case _, ok := <-events:
		if ok {
			t.Fatal("expected event channel to close after process_died")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event channel close")
	}
}
