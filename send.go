package pi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var defaultRequestTimeout = 2 * time.Minute

func (client *Client) send(ctx context.Context, command rpcCommand) (rpcResponse, error) {
	commandType, err := commandTypeOf(command)
	if err != nil {
		return rpcResponse{}, err
	}

	ctx, cancel, err := withDefaultRequestTimeout(ctx)
	if err != nil {
		return rpcResponse{}, err
	}
	defer cancel()

	if err := client.currentProcessError(); err != nil {
		return rpcResponse{}, err
	}

	requestID := client.nextRequestID()
	payloadCommand := cloneCommand(command)
	payloadCommand["id"] = requestID

	payload, err := json.Marshal(payloadCommand)
	if err != nil {
		return rpcResponse{}, err
	}

	responseChan := make(chan rpcResponse, 1)
	if err := client.requests.register(requestID, responseChan); err != nil {
		return rpcResponse{}, err
	}

	client.writeLock.Lock()
	_, writeErr := client.stdin.Write(append(payload, '\n'))
	client.writeLock.Unlock()
	if writeErr != nil {
		client.requests.drop(requestID)
		if err := client.terminalError(); err != nil {
			return rpcResponse{}, err
		}
		return rpcResponse{}, fmt.Errorf("write %s command: %w", commandType, writeErr)
	}

	select {
	case <-ctx.Done():
		client.requests.drop(requestID)
		return rpcResponse{}, ctx.Err()
	case <-client.closed:
		client.requests.drop(requestID)
		if err := client.terminalError(); err != nil {
			return rpcResponse{}, err
		}
		return rpcResponse{}, ErrClientClosed
	case response, ok := <-responseChan:
		if !ok {
			if err := client.terminalError(); err != nil {
				return rpcResponse{}, err
			}
			return rpcResponse{}, fmt.Errorf("%w: closed response channel for request %s", ErrProtocolViolation, requestID)
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

func commandTypeOf(command rpcCommand) (string, error) {
	if command == nil {
		return "", errors.New("command is required")
	}
	commandType, ok := command["type"].(string)
	if !ok || strings.TrimSpace(commandType) == "" {
		return "", errors.New("command type is required")
	}
	return strings.TrimSpace(commandType), nil
}

func rpcErrorFromResponse(response rpcResponse) error {
	message := strings.TrimSpace(response.Error)
	if message == ErrProcessDied.Error() {
		return ErrProcessDied
	}
	return &RPCError{RequestID: response.ID, Command: response.Command, Message: message}
}

func cloneCommand(command rpcCommand) rpcCommand {
	cloned := make(rpcCommand, len(command)+1)
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
