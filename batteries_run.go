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
	detailed, err := client.RunDetailed(ctx, request)
	if err != nil {
		return RunResult{}, err
	}
	return RunResult{Text: detailed.Outcome.Text, Usage: detailed.Outcome.Usage}, nil
}

func (client *Client) RunDetailed(ctx context.Context, request PromptRequest) (RunDetailedResult, error) {
	if ctx == nil {
		return RunDetailedResult{}, ErrNilContext
	}
	if !client.runInProgress.CompareAndSwap(false, true) {
		return RunDetailedResult{}, ErrRunInProgress
	}
	defer client.runInProgress.Store(false)

	events, cancel, err := client.Subscribe(SubscriptionPolicy{Buffer: 256, Mode: SubscriptionModeRing})
	if err != nil {
		return RunDetailedResult{}, err
	}
	defer cancel()

	command, err := promptCommand(request)
	if err != nil {
		return RunDetailedResult{}, err
	}
	promptResponse, err := client.send(ctx, command)
	if err != nil {
		return RunDetailedResult{}, err
	}

	return client.waitForRunDetailed(ctx, events, promptResponse.ID)
}

func (client *Client) waitForRunDetailed(ctx context.Context, events <-chan Event, promptRequestID string) (RunDetailedResult, error) {
	result := RunDetailedResult{}
	for {
		select {
		case <-ctx.Done():
			client.abortRunBestEffort()
			return RunDetailedResult{}, ctx.Err()
		case event, ok := <-events:
			if !ok {
				if err := client.terminalError(); err != nil {
					return RunDetailedResult{}, err
				}
				return RunDetailedResult{}, fmt.Errorf("%w: event stream closed", ErrProtocolViolation)
			}

			if event.Type == EventTypeProcessDied {
				if err := client.currentProcessError(); err != nil {
					return RunDetailedResult{}, err
				}
				return RunDetailedResult{}, ErrProcessDied
			}

			if err, handled := asyncPromptFailure(event, promptRequestID); handled {
				if err != nil {
					return RunDetailedResult{}, err
				}
				continue
			}

			switch event.Type {
			case EventTypeAutoCompactionStart:
				parsed, err := DecodeAutoCompactionStart(event.Raw)
				if err == nil {
					result.AutoCompactionStart = &parsed
				}
				continue
			case EventTypeAutoCompactionEnd:
				parsed, err := DecodeAutoCompactionEnd(event.Raw)
				if err == nil {
					result.AutoCompactionEnd = &parsed
				}
				continue
			case EventTypeAutoRetryStart:
				parsed, err := DecodeAutoRetryStart(event.Raw)
				if err == nil {
					result.AutoRetryStart = &parsed
				}
				continue
			case EventTypeAutoRetryEnd:
				parsed, err := DecodeAutoRetryEnd(event.Raw)
				if err == nil {
					result.AutoRetryEnd = &parsed
				}
				continue
			case EventTypeAgentEnd:
				outcome, err := DecodeTerminalOutcome(event.Raw)
				if err != nil {
					return RunDetailedResult{}, err
				}
				result.Outcome = outcome
				return result, nil
			default:
				continue
			}
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
