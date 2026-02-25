package pi

import "sync"

type eventHub struct {
	mu          sync.Mutex
	subscribers map[*subscription]struct{}
	closed      bool
}

func newEventHub() *eventHub {
	return &eventHub{subscribers: map[*subscription]struct{}{}}
}

func (hub *eventHub) subscribe(policy SubscriptionPolicy) (<-chan Event, func(), error) {
	sub, err := newSubscription(policy)
	if err != nil {
		return nil, nil, err
	}

	hub.mu.Lock()
	if hub.closed {
		hub.mu.Unlock()
		return nil, nil, ErrClientClosed
	}
	hub.subscribers[sub] = struct{}{}
	hub.mu.Unlock()

	go sub.run()

	var cancelOnce sync.Once
	cancel := func() {
		cancelOnce.Do(func() {
			hub.mu.Lock()
			delete(hub.subscribers, sub)
			hub.mu.Unlock()
			sub.close()
		})
	}

	return sub.out, cancel, nil
}

func (hub *eventHub) publish(event Event) {
	subscribers := hub.snapshot()
	publishToSubscribers(subscribers, event)
}

func (hub *eventHub) processDied(event Event) {
	subscribers := hub.closeAndSnapshot()
	publishToSubscribers(subscribers, event)
	for sub := range subscribers {
		sub.close()
	}
}

func (hub *eventHub) close() {
	subscribers := hub.closeAndSnapshot()
	for sub := range subscribers {
		sub.close()
	}
}

func (hub *eventHub) snapshot() map[*subscription]struct{} {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	copy := make(map[*subscription]struct{}, len(hub.subscribers))
	for sub := range hub.subscribers {
		copy[sub] = struct{}{}
	}
	return copy
}

func (hub *eventHub) closeAndSnapshot() map[*subscription]struct{} {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	if hub.closed {
		return map[*subscription]struct{}{}
	}
	hub.closed = true
	copy := make(map[*subscription]struct{}, len(hub.subscribers))
	for sub := range hub.subscribers {
		copy[sub] = struct{}{}
	}
	hub.subscribers = map[*subscription]struct{}{}
	return copy
}

func publishToSubscribers(subscribers map[*subscription]struct{}, event Event) {
	for sub := range subscribers {
		dropped := sub.enqueue(event)
		if dropped && sub.policy.EmitDropEvent && event.Type != EventTypeSubscriptionDrop {
			sub.enqueueSystem(newSubscriptionDropEvent(sub.policy.Mode, event.Type))
		}
	}
}
