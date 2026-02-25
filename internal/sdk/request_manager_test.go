package sdk

import (
	"errors"
	"testing"
	"time"

	"github.com/joshp123/pi-golang/internal/rpc"
	"github.com/joshp123/pi-golang/internal/transport"
)

func TestRequestManagerResolvesPendingResponse(t *testing.T) {
	manager := transport.NewRequestManager(ErrClientClosed)

	responseChan := make(chan rpc.Response, 1)
	if err := manager.Register("req-1", responseChan); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if resolved := manager.Resolve(rpc.Response{ID: "req-1", Command: rpc.CommandPrompt, Success: true}); !resolved {
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
	manager := transport.NewRequestManager(ErrClientClosed)

	manager.Close(nil)

	err := manager.Register("req-2", make(chan rpc.Response, 1))
	if !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed, got %v", err)
	}
}

func TestRequestManagerProcessDiedRejectsNewRequests(t *testing.T) {
	manager := transport.NewRequestManager(ErrClientClosed)

	manager.MarkProcessDied(ErrProcessDied)

	err := manager.Register("req-3", make(chan rpc.Response, 1))
	if !errors.Is(err, ErrProcessDied) {
		t.Fatalf("expected ErrProcessDied, got %v", err)
	}
}

func TestRequestManagerProcessDiedClosesPendingRequests(t *testing.T) {
	manager := transport.NewRequestManager(ErrClientClosed)
	responseChan := make(chan rpc.Response, 1)
	if err := manager.Register("req-4", responseChan); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	manager.MarkProcessDied(ErrProcessDied)

	select {
	case _, ok := <-responseChan:
		if ok {
			t.Fatal("expected pending response channel to be closed")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for pending response channel close")
	}

	if err := manager.CurrentError(); !errors.Is(err, ErrProcessDied) {
		t.Fatalf("expected ErrProcessDied current error, got %v", err)
	}
}
