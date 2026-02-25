package pi

import (
	"context"
	"fmt"
	"time"
)

var defaultRunAbortTimeout = 2 * time.Second

// Batteries layer: higher-level helpers built on top of thin RPC methods.
//
// Run mechanics:
//  1. Send one prompt command.
//  2. Consume events until agent_end or terminal failure.
//  3. If ctx is cancelled while waiting, best-effort abort then return ctx error.

func (client *Client) Run(ctx context.Context, request PromptRequest) (RunResult, error) {
	if ctx == nil {
		return RunResult{}, ErrNilContext
	}
	if !client.runInProgress.CompareAndSwap(false, true) {
		return RunResult{}, ErrRunInProgress
	}
	defer client.runInProgress.Store(false)

	events, cancel, err := client.Subscribe(SubscriptionPolicy{Buffer: 256, Mode: SubscriptionModeRing})
	if err != nil {
		return RunResult{}, err
	}
	defer cancel()

	command, err := promptCommand(request)
	if err != nil {
		return RunResult{}, err
	}
	promptResponse, err := client.send(ctx, command)
	if err != nil {
		return RunResult{}, err
	}

	return client.waitForRunCompletion(ctx, events, promptResponse.ID)
}

func (client *Client) waitForRunCompletion(ctx context.Context, events <-chan Event, promptRequestID string) (RunResult, error) {
	for {
		select {
		case <-ctx.Done():
			client.abortRunBestEffort()
			return RunResult{}, ctx.Err()
		case event, ok := <-events:
			if !ok {
				if err := client.terminalError(); err != nil {
					return RunResult{}, err
				}
				return RunResult{}, fmt.Errorf("%w: event stream closed", ErrProtocolViolation)
			}

			if event.Type == EventTypeProcessDied {
				if err := client.currentProcessError(); err != nil {
					return RunResult{}, err
				}
				return RunResult{}, ErrProcessDied
			}

			if err, handled := asyncPromptFailure(event, promptRequestID); handled {
				if err != nil {
					return RunResult{}, err
				}
				continue
			}

			if event.Type != EventTypeAgentEnd {
				continue
			}
			return extractRunResult(event)
		}
	}
}

func asyncPromptFailure(event Event, promptRequestID string) (error, bool) {
	if event.Type != eventTypeResponse || promptRequestID == "" {
		return nil, false
	}
	response, err := decodeRPCResponse(event.Raw)
	if err != nil {
		return nil, true
	}
	if response.ID != promptRequestID || response.Command != rpcCommandPrompt || response.Success {
		return nil, true
	}
	return rpcErrorFromResponse(response), true
}

func (client *Client) abortRunBestEffort() {
	if err := client.terminalError(); err != nil {
		return
	}
	abortCtx, cancel := context.WithTimeout(context.Background(), defaultRunAbortTimeout)
	defer cancel()
	_ = client.Abort(abortCtx)
}
