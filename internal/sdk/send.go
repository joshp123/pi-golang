package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/joshp123/pi-golang/internal/rpc"
)

var defaultRequestTimeout = 2 * time.Minute

func (client *Client) send(ctx context.Context, command rpc.Command) (rpc.Response, error) {
	commandType, err := commandTypeOf(command)
	if err != nil {
		return rpc.Response{}, err
	}

	ctx, cancel, err := withDefaultRequestTimeout(ctx)
	if err != nil {
		return rpc.Response{}, err
	}
	defer cancel()

	if err := client.currentProcessError(); err != nil {
		return rpc.Response{}, err
	}

	requestID := client.nextRequestID()
	payloadCommand := cloneCommand(command)
	payloadCommand["id"] = requestID

	payload, err := json.Marshal(payloadCommand)
	if err != nil {
		return rpc.Response{}, err
	}

	responseChan := make(chan rpc.Response, 1)
	if err := client.requests.Register(requestID, responseChan); err != nil {
		return rpc.Response{}, err
	}

	client.writeLock.Lock()
	_, writeErr := client.stdin.Write(append(payload, '\n'))
	client.writeLock.Unlock()
	if writeErr != nil {
		client.requests.Drop(requestID)
		if err := client.terminalError(); err != nil {
			return rpc.Response{}, err
		}
		return rpc.Response{}, fmt.Errorf("write %s command: %w", commandType, writeErr)
	}

	select {
	case <-ctx.Done():
		client.requests.Drop(requestID)
		return rpc.Response{}, ctx.Err()
	case <-client.closed:
		client.requests.Drop(requestID)
		if err := client.terminalError(); err != nil {
			return rpc.Response{}, err
		}
		return rpc.Response{}, ErrClientClosed
	case response, ok := <-responseChan:
		if !ok {
			if err := client.terminalError(); err != nil {
				return rpc.Response{}, err
			}
			return rpc.Response{}, fmt.Errorf("%w: closed response channel for request %s", ErrProtocolViolation, requestID)
		}
		if !response.Success {
			return response, rpcErrorFromResponse(response)
		}
		return response, nil
	}
}

func withDefaultRequestTimeout(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if ctx == nil {
		return nil, nil, ErrNilContext
	}
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}, nil
	}
	timedCtx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	return timedCtx, cancel, nil
}

func commandTypeOf(command rpc.Command) (string, error) {
	if command == nil {
		return "", errors.New("command is required")
	}
	commandType, ok := command["type"].(string)
	if !ok || strings.TrimSpace(commandType) == "" {
		return "", errors.New("command type is required")
	}
	return strings.TrimSpace(commandType), nil
}

func rpcErrorFromResponse(response rpc.Response) error {
	message := strings.TrimSpace(response.Error)
	if message == ErrProcessDied.Error() {
		return ErrProcessDied
	}
	return &RPCError{RequestID: response.ID, Command: response.Command, Message: message}
}

func cloneCommand(command rpc.Command) rpc.Command {
	cloned := make(rpc.Command, len(command)+1)
	for key, value := range command {
		cloned[key] = value
	}
	return cloned
}

func (client *Client) isClosed() bool {
	select {
	case <-client.closed:
		return true
	default:
		return false
	}
}

func (client *Client) terminalError() error {
	if err := client.currentProcessError(); err != nil {
		return err
	}
	if client.isClosed() {
		return ErrClientClosed
	}
	return nil
}
