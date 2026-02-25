package pi

import (
	"errors"
	"testing"
	"time"
)

func TestRequestManagerResolvesPendingResponse(t *testing.T) {
	manager := newRequestManager()

	responseChan := make(chan rpcResponse, 1)
	if err := manager.register("req-1", responseChan); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if resolved := manager.resolve(rpcResponse{ID: "req-1", Command: rpcCommandPrompt, Success: true}); !resolved {
		t.Fatal("expected resolve=true")
	}

	select {
	case response, ok := <-responseChan:
		if !ok {
			t.Fatal("response channel closed before delivery")
		}
		if response.ID != "req-1" {
			t.Fatalf("unexpected response id: %s", response.ID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for resolved response")
	}

	if _, ok := <-responseChan; ok {
		t.Fatal("expected resolved response channel to close")
	}
}

func TestRequestManagerCloseRejectsNewRequests(t *testing.T) {
	manager := newRequestManager()

	manager.close(nil)

	err := manager.register("req-2", make(chan rpcResponse, 1))
	if !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed, got %v", err)
	}
}

func TestRequestManagerProcessDiedRejectsNewRequests(t *testing.T) {
	manager := newRequestManager()

	manager.markProcessDied(ErrProcessDied)

	err := manager.register("req-3", make(chan rpcResponse, 1))
	if !errors.Is(err, ErrProcessDied) {
		t.Fatalf("expected ErrProcessDied, got %v", err)
	}
}

func TestRequestManagerProcessDiedClosesPendingRequests(t *testing.T) {
	manager := newRequestManager()
	responseChan := make(chan rpcResponse, 1)
	if err := manager.register("req-4", responseChan); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	manager.markProcessDied(ErrProcessDied)

	select {
	case _, ok := <-responseChan:
		if ok {
			t.Fatal("expected pending response channel to be closed")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for pending response channel close")
	}

	if err := manager.currentError(); !errors.Is(err, ErrProcessDied) {
		t.Fatalf("expected ErrProcessDied current error, got %v", err)
	}
}
