package transport

import (
	"github.com/joshp123/pi-golang/internal/rpc"
	"github.com/joshp123/pi-golang/internal/runtime"
)

// RequestManager owns pending request channels + terminal process error state.
type RequestManager struct {
	registry *runtime.PendingRegistry[rpc.Response]
}

func NewRequestManager(closedErr error) *RequestManager {
	return &RequestManager{registry: runtime.NewPendingRegistry[rpc.Response](closedErr)}
}

func (manager *RequestManager) Register(requestID string, response chan rpc.Response) error {
	return manager.registry.Register(requestID, response)
}

func (manager *RequestManager) Drop(requestID string) {
	manager.registry.Drop(requestID)
}

func (manager *RequestManager) Resolve(response rpc.Response) bool {
	return manager.registry.Resolve(response.ID, response)
}

func (manager *RequestManager) CurrentError() error {
	return manager.registry.CurrentError()
}

func (manager *RequestManager) MarkProcessDied(err error) {
	manager.registry.MarkProcessDied(err)
}

func (manager *RequestManager) Close(err error) {
	manager.registry.Close(err)
}
