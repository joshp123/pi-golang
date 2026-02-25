package pi

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrProcessDied indicates the underlying pi RPC process terminated unexpectedly.
	ErrProcessDied = errors.New("pi process died")
	// ErrClientClosed indicates the client was closed by the caller.
	ErrClientClosed = errors.New("pi client closed")
	// ErrProtocolViolation indicates invalid output from the pi RPC process.
	ErrProtocolViolation = errors.New("pi rpc protocol violation")
	// ErrNilContext indicates a required context argument was nil.
	ErrNilContext = errors.New("context is required")
	// ErrRunInProgress indicates Run was called while another Run was already active.
	ErrRunInProgress = errors.New("run already in progress")
	// ErrInvalidSubscriptionPolicy indicates an unsupported subscription mode or buffer.
	ErrInvalidSubscriptionPolicy = errors.New("invalid subscription policy")
)

// RPCError is returned when pi responds with success=false for a command.
type RPCError struct {
	RequestID string
	Command   string
	Message   string
}

func (err *RPCError) Error() string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Message)
	if message == "" {
		message = "rpc command failed"
	}
	if err.Command == "" {
		return message
	}
	if err.RequestID == "" {
		return fmt.Sprintf("rpc %s failed: %s", err.Command, message)
	}
	return fmt.Sprintf("rpc %s (%s) failed: %s", err.Command, err.RequestID, message)
}
