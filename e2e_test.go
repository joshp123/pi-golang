package pi

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestClientTypedWrappersE2E(t *testing.T) {
	setupFakePI(t, "happy")

	client, err := StartOneShot(DefaultOneShotOptions())
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	if _, err := client.GetState(nil); !errors.Is(err, ErrNilContext) {
		t.Fatalf("expected ErrNilContext from GetState(nil), got %v", err)
	}
	if err := client.Abort(nil); !errors.Is(err, ErrNilContext) {
		t.Fatalf("expected ErrNilContext from Abort(nil), got %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, err := client.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if state.SessionID != "session-123" {
		t.Fatalf("unexpected session id: %s", state.SessionID)
	}
	if state.ContextWindow != 200000 {
		t.Fatalf("unexpected context window: %d", state.ContextWindow)
	}
	if !state.AutoCompactionEnabled {
		t.Fatal("expected auto compaction to be enabled")
	}

	cancelled, err := client.NewSession(ctx, "cancel-parent")
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	if !cancelled {
		t.Fatal("expected NewSession to return cancelled=true")
	}

	compactResult, err := client.Compact(ctx, "focus on key changes")
	if err != nil {
		t.Fatalf("Compact failed: %v", err)
	}
	if compactResult.Summary != "compacted" {
		t.Fatalf("unexpected compact summary: %q", compactResult.Summary)
	}

	if _, err := client.Compact(ctx, "force-error"); err == nil {
		t.Fatal("expected compact error for force-error instructions")
	}

	runResult, err := client.Run(ctx, PromptRequest{Message: "hello"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if runResult.Text != "hello from helper" {
		t.Fatalf("unexpected run text: %q", runResult.Text)
	}

	if err := client.Abort(ctx); err != nil {
		t.Fatalf("Abort failed: %v", err)
	}
}

func TestProcessDeathFailsPendingAndClosesSubscribers(t *testing.T) {
	setupFakePI(t, "die_on_prompt")

	client, err := StartOneShot(DefaultOneShotOptions())
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	events, cancelEvents, err := client.Subscribe(SubscriptionPolicy{Buffer: 8, Mode: SubscriptionModeRing})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer cancelEvents()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Prompt(ctx, PromptRequest{Message: "boom"})
	if !errors.Is(err, ErrProcessDied) {
		t.Fatalf("expected ErrProcessDied, got: %v", err)
	}

	event := readEventOrFail(t, events)
	if event.Type != EventTypeProcessDied {
		t.Fatalf("expected %s event, got %s", EventTypeProcessDied, event.Type)
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

	client, err := StartOneShot(DefaultOneShotOptions())
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.Run(ctx, PromptRequest{Message: "hello"})
	if err == nil {
		t.Fatal("expected Run to fail")
	}
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected RPCError, got %T: %v", err, err)
	}
	if rpcErr.Command != rpcCommandPrompt {
		t.Fatalf("expected prompt command error, got %q", rpcErr.Command)
	}
}

func TestRunRejectsConcurrentRun(t *testing.T) {
	setupFakePI(t, "slow_run")

	client, err := StartOneShot(DefaultOneShotOptions())
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	runErr := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := client.Run(ctx, PromptRequest{Message: "first"})
		runErr <- err
	}()

	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = client.Run(ctx, PromptRequest{Message: "second"})
	if !errors.Is(err, ErrRunInProgress) {
		t.Fatalf("expected ErrRunInProgress, got %v", err)
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

	client, err := StartOneShot(DefaultOneShotOptions())
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	runResult := make(chan RunResult, 1)
	runErr := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		result, err := client.Run(ctx, PromptRequest{Message: "start"})
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

	client, err := StartOneShot(DefaultOneShotOptions())
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		_, runErrValue := client.Run(ctx, PromptRequest{Message: "start"})
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

func TestSendStillReturnsResponseWhenBlockSubscriberIsNotConsuming(t *testing.T) {
	setupFakePI(t, "flood_before_response")

	client, err := StartOneShot(DefaultOneShotOptions())
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	_, cancel, err := client.Subscribe(SubscriptionPolicy{Buffer: 1, Mode: SubscriptionModeBlock})
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

	client, err := StartOneShot(DefaultOneShotOptions())
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
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
		if !errors.Is(sendErr, ErrClientClosed) {
			t.Fatalf("expected ErrClientClosed, got %v", sendErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pending request to return")
	}
}
