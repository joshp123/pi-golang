package pi

import "github.com/joshp123/pi-golang/internal/runtime"

// requestManager owns pending request channels + terminal process error state.
// Root wrapper around internal/runtime PendingRegistry.

type requestManager struct {
	registry *runtime.PendingRegistry[rpcResponse]
}

func newRequestManager() *requestManager {
	return &requestManager{registry: runtime.NewPendingRegistry[rpcResponse](ErrClientClosed)}
}

func (manager *requestManager) register(requestID string, response chan rpcResponse) error {
	return manager.registry.Register(requestID, response)
}

func (manager *requestManager) drop(requestID string) {
	manager.registry.Drop(requestID)
}

func (manager *requestManager) resolve(response rpcResponse) bool {
	return manager.registry.Resolve(response.ID, response)
}

func (manager *requestManager) currentError() error {
	return manager.registry.CurrentError()
}

func (manager *requestManager) markProcessDied(err error) {
	manager.registry.MarkProcessDied(err)
}

func (manager *requestManager) close(err error) {
	manager.registry.Close(err)
}
