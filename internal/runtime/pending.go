package runtime

import "sync"

// PendingRegistry tracks in-flight request channels and terminal process state.
// Generic over response payload type so root package can keep wire types private.
type PendingRegistry[T any] struct {
	mu         sync.Mutex
	pending    map[string]chan T
	processErr error
	closed     bool
	closedErr  error
}

func NewPendingRegistry[T any](closedErr error) *PendingRegistry[T] {
	return &PendingRegistry[T]{
		pending:   map[string]chan T{},
		closedErr: closedErr,
	}
}

func (registry *PendingRegistry[T]) Register(requestID string, response chan T) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if registry.processErr != nil {
		return registry.processErr
	}
	if registry.closed {
		return registry.closedErr
	}
	registry.pending[requestID] = response
	return nil
}

func (registry *PendingRegistry[T]) Drop(requestID string) {
	responseChan := registry.takePending(requestID)
	if responseChan == nil {
		return
	}
	close(responseChan)
}

func (registry *PendingRegistry[T]) Resolve(requestID string, response T) bool {
	responseChan := registry.takePending(requestID)
	if responseChan == nil {
		return false
	}
	responseChan <- response
	close(responseChan)
	return true
}

func (registry *PendingRegistry[T]) CurrentError() error {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	return registry.processErr
}

func (registry *PendingRegistry[T]) MarkProcessDied(err error) {
	for _, responseChan := range registry.failPending(err, false) {
		close(responseChan)
	}
}

func (registry *PendingRegistry[T]) Close(err error) {
	for _, responseChan := range registry.failPending(err, true) {
		close(responseChan)
	}
}

func (registry *PendingRegistry[T]) takePending(requestID string) chan T {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	responseChan := registry.pending[requestID]
	if responseChan == nil {
		return nil
	}
	delete(registry.pending, requestID)
	return responseChan
}

func (registry *PendingRegistry[T]) failPending(err error, closed bool) []chan T {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if registry.processErr == nil && err != nil {
		registry.processErr = err
	}
	if closed {
		registry.closed = true
	}

	pending := make([]chan T, 0, len(registry.pending))
	for requestID, responseChan := range registry.pending {
		pending = append(pending, responseChan)
		delete(registry.pending, requestID)
	}
	return pending
}
