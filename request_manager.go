package pi

import "sync"

// requestManager owns pending request channels + terminal process error state.
// Intentionally simple: mutex + map, no supervisor goroutine.

type requestManager struct {
	mu         sync.Mutex
	pending    map[string]chan rpcResponse
	processErr error
	closed     bool
}

func newRequestManager() *requestManager {
	return &requestManager{pending: map[string]chan rpcResponse{}}
}

func (manager *requestManager) register(requestID string, response chan rpcResponse) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.processErr != nil {
		return manager.processErr
	}
	if manager.closed {
		return ErrClientClosed
	}
	manager.pending[requestID] = response
	return nil
}

func (manager *requestManager) drop(requestID string) {
	manager.closePendingByID(requestID)
}

func (manager *requestManager) resolve(response rpcResponse) bool {
	responseChan := manager.takePending(response.ID)
	if responseChan == nil {
		return false
	}
	responseChan <- response
	close(responseChan)
	return true
}

func (manager *requestManager) currentError() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.processErr
}

func (manager *requestManager) markProcessDied(err error) {
	for _, responseChan := range manager.failPending(err, false) {
		close(responseChan)
	}
}

func (manager *requestManager) close(err error) {
	for _, responseChan := range manager.failPending(err, true) {
		close(responseChan)
	}
}

func (manager *requestManager) closePendingByID(requestID string) {
	responseChan := manager.takePending(requestID)
	if responseChan == nil {
		return
	}
	close(responseChan)
}

func (manager *requestManager) takePending(requestID string) chan rpcResponse {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	responseChan := manager.pending[requestID]
	if responseChan == nil {
		return nil
	}
	delete(manager.pending, requestID)
	return responseChan
}

func (manager *requestManager) failPending(err error, closed bool) []chan rpcResponse {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.processErr == nil && err != nil {
		manager.processErr = err
	}
	if closed {
		manager.closed = true
	}

	pending := make([]chan rpcResponse, 0, len(manager.pending))
	for requestID, responseChan := range manager.pending {
		pending = append(pending, responseChan)
		delete(manager.pending, requestID)
	}
	return pending
}
