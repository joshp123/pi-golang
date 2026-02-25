package sdk_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sdk "github.com/joshp123/pi-golang/internal/sdk"
)

func TestProcessDeathFailsPendingAndClosesSubscribers(t *testing.T) {
	setupFakePI(t, "die_on_prompt")

	client, err := sdk.StartOneShot(testOneShotOptions())
	if err != nil {
		t.Fatalf("sdk.StartOneShot failed: %v", err)
	}
	defer client.Close()

	events, cancelEvents, err := client.Subscribe(sdk.SubscriptionPolicy{Buffer: 8, Mode: sdk.SubscriptionModeRing})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer cancelEvents()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Prompt(ctx, sdk.PromptRequest{Message: "boom"})
	if !errors.Is(err, sdk.ErrProcessDied) {
		t.Fatalf("expected sdk.ErrProcessDied, got: %v", err)
	}

	event := readEventOrFail(t, events)
	if event.Type != sdk.EventTypeProcessDied {
		t.Fatalf("expected %s event, got %s", sdk.EventTypeProcessDied, event.Type)
	}

	select {
	case _, ok := <-events:
		if ok {
			t.Fatal("expected subscriber channel to close")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber channel close")
	}
}

func TestRunPropagatesPromptAsyncFailure(t *testing.T) {
	setupFakePI(t, "prompt_async_error")

	client, err := sdk.StartOneShot(testOneShotOptions())
	if err != nil {
		t.Fatalf("sdk.StartOneShot failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.Run(ctx, sdk.PromptRequest{Message: "hello"})
	if err == nil {
		t.Fatal("expected Run to fail")
	}
	var rpcErr *sdk.RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected sdk.RPCError, got %T: %v", err, err)
	}
	if rpcErr.Command != "prompt" {
		t.Fatalf("expected prompt command error, got %q", rpcErr.Command)
	}
}

func TestRunRejectsConcurrentRun(t *testing.T) {
	setupFakePI(t, "slow_run")

	client, err := sdk.StartOneShot(testOneShotOptions())
	if err != nil {
		t.Fatalf("sdk.StartOneShot failed: %v", err)
	}
	defer client.Close()

	runErr := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := client.Run(ctx, sdk.PromptRequest{Message: "first"})
		runErr <- err
	}()

	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = client.Run(ctx, sdk.PromptRequest{Message: "second"})
	if !errors.Is(err, sdk.ErrRunInProgress) {
		t.Fatalf("expected sdk.ErrRunInProgress, got %v", err)
	}

	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("first run failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for first run")
	}
}

func TestAbortInterruptsRun(t *testing.T) {
	setupFakePI(t, "abort_run")

	client, err := sdk.StartOneShot(testOneShotOptions())
	if err != nil {
		t.Fatalf("sdk.StartOneShot failed: %v", err)
	}
	defer client.Close()

	runResult := make(chan sdk.RunResult, 1)
	runErr := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		result, err := client.Run(ctx, sdk.PromptRequest{Message: "start"})
		runResult <- result
		runErr <- err
	}()

	time.Sleep(50 * time.Millisecond)
	abortCtx, cancelAbort := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelAbort()
	if err := client.Abort(abortCtx); err != nil {
		t.Fatalf("Abort failed: %v", err)
	}

	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("Run failed after Abort: %v", err)
		}
		result := <-runResult
		if result.Text != "aborted by helper" {
			t.Fatalf("unexpected run text after Abort: %q", result.Text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for aborted run")
	}
}

func TestRunContextCancellationBestEffortAborts(t *testing.T) {
	setupFakePI(t, "run_ctx_cancel_aborts")

	client, err := sdk.StartOneShot(testOneShotOptions())
	if err != nil {
		t.Fatalf("sdk.StartOneShot failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		_, runErrValue := client.Run(ctx, sdk.PromptRequest{Message: "start"})
		runErr <- runErrValue
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()

	err = <-runErr
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation from Run, got %v", err)
	}

	stateCtx, cancelState := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelState()
	state, err := client.GetState(stateCtx)
	if err != nil {
		t.Fatalf("GetState after cancelled Run failed: %v", err)
	}
	if state.SessionID != "abort-observed" {
		t.Fatalf("expected helper to observe Abort after Run cancel, got session id %q", state.SessionID)
	}
}

func TestRunDetailedCapturesCompactionAndRetrySignals(t *testing.T) {
	setupFakePI(t, "run_detailed_signals")

	client, err := sdk.StartOneShot(testOneShotOptions())
	if err != nil {
		t.Fatalf("sdk.StartOneShot failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := client.RunDetailed(ctx, sdk.PromptRequest{Message: "start"})
	if err != nil {
		t.Fatalf("RunDetailed failed: %v", err)
	}
	if result.Outcome.Text != "after compaction" {
		t.Fatalf("unexpected outcome text: %q", result.Outcome.Text)
	}
	if result.AutoCompactionStart == nil || result.AutoCompactionStart.Reason != "overflow" {
		t.Fatalf("expected auto_compaction_start overflow, got %+v", result.AutoCompactionStart)
	}
	if result.AutoCompactionEnd == nil || !result.AutoCompactionEnd.WillRetry {
		t.Fatalf("expected auto_compaction_end with willRetry=true, got %+v", result.AutoCompactionEnd)
	}
	if result.AutoRetryStart == nil || result.AutoRetryStart.Attempt != 1 {
		t.Fatalf("expected auto_retry_start attempt=1, got %+v", result.AutoRetryStart)
	}
	if result.AutoRetryEnd == nil || !result.AutoRetryEnd.Success {
		t.Fatalf("expected auto_retry_end success=true, got %+v", result.AutoRetryEnd)
	}
}

func TestSendStillReturnsResponseWhenBlockSubscriberIsNotConsuming(t *testing.T) {
	setupFakePI(t, "flood_before_response")

	client, err := sdk.StartOneShot(testOneShotOptions())
	if err != nil {
		t.Fatalf("sdk.StartOneShot failed: %v", err)
	}
	defer client.Close()

	_, cancel, err := client.Subscribe(sdk.SubscriptionPolicy{Buffer: 1, Mode: sdk.SubscriptionModeBlock})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer cancel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelCtx()

	state, err := client.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if state.SessionID != "session-123" {
		t.Fatalf("unexpected session id: %s", state.SessionID)
	}
}

func TestCloseUnblocksPendingRequest(t *testing.T) {
	setupFakePI(t, "never_respond")

	client, err := sdk.StartOneShot(testOneShotOptions())
	if err != nil {
		t.Fatalf("sdk.StartOneShot failed: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		_, sendErr := client.GetState(context.Background())
		errCh <- sendErr
	}()

	time.Sleep(100 * time.Millisecond)
	if err := client.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	select {
	case sendErr := <-errCh:
		if !errors.Is(sendErr, sdk.ErrClientClosed) {
			t.Fatalf("expected ErrClientClosed, got %v", sendErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pending request to return")
	}
}

func testOneShotOptions() sdk.OneShotOptions {
	opts := sdk.DefaultOneShotOptions()
	opts.Auth.Anthropic.APIKey = sdk.Credential{Value: "test-key"}
	return opts
}
