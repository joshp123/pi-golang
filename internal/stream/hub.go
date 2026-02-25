package stream

import "sync"

type Hub[T any] struct {
	mu            sync.Mutex
	subscribers   map[*subscription[T]]struct{}
	closed        bool
	closedErr     error
	eventType     func(T) string
	dropEventType string
	newDropEvent  func(Mode, string) T
}

func NewHub[T any](closedErr error, eventType func(T) string, dropEventType string, newDropEvent func(Mode, string) T) *Hub[T] {
	return &Hub[T]{
		subscribers:   map[*subscription[T]]struct{}{},
		closedErr:     closedErr,
		eventType:     eventType,
		dropEventType: dropEventType,
		newDropEvent:  newDropEvent,
	}
}

func (hub *Hub[T]) Subscribe(policy Policy) (<-chan T, func(), error) {
	sub := newSubscription[T](policy)

	hub.mu.Lock()
	if hub.closed {
		hub.mu.Unlock()
		return nil, nil, hub.closedErr
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

func (hub *Hub[T]) Publish(event T) {
	subscribers := hub.snapshot()
	hub.publishToSubscribers(subscribers, event)
}

func (hub *Hub[T]) ProcessDied(event T) {
	subscribers := hub.closeAndSnapshot()
	hub.publishToSubscribers(subscribers, event)
	for sub := range subscribers {
		sub.close()
	}
}

func (hub *Hub[T]) Close() {
	subscribers := hub.closeAndSnapshot()
	for sub := range subscribers {
		sub.close()
	}
}

func (hub *Hub[T]) snapshot() map[*subscription[T]]struct{} {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	copy := make(map[*subscription[T]]struct{}, len(hub.subscribers))
	for sub := range hub.subscribers {
		copy[sub] = struct{}{}
	}
	return copy
}

func (hub *Hub[T]) closeAndSnapshot() map[*subscription[T]]struct{} {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	if hub.closed {
		return map[*subscription[T]]struct{}{}
	}
	hub.closed = true
	copy := make(map[*subscription[T]]struct{}, len(hub.subscribers))
	for sub := range hub.subscribers {
		copy[sub] = struct{}{}
	}
	hub.subscribers = map[*subscription[T]]struct{}{}
	return copy
}

func (hub *Hub[T]) publishToSubscribers(subscribers map[*subscription[T]]struct{}, event T) {
	for sub := range subscribers {
		dropped := sub.enqueue(event)
		if dropped && sub.policy.EmitDropEvent && hub.newDropEvent != nil && hub.eventType != nil {
			if eventType := hub.eventType(event); eventType != "" && eventType != hub.dropEventType {
				sub.enqueueSystem(hub.newDropEvent(sub.policy.Mode, eventType))
			}
		}
	}
}
