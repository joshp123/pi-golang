package sdk_test

import (
	"encoding/json"
	"testing"

	sdk "github.com/joshp123/pi-golang/internal/sdk"
)

func TestDecodeTerminalOutcomeCompleted(t *testing.T) {
	raw := json.RawMessage(`{
		"type":"agent_end",
		"messages":[
			{"role":"user","content":"hi"},
			{"role":"assistant","content":[{"type":"text","text":"done"}],"stopReason":"stop","usage":{"input":1,"output":2}}
		]
	}`)

	outcome, err := sdk.DecodeTerminalOutcome(raw)
	if err != nil {
		t.Fatalf("DecodeTerminalOutcome returned error: %v", err)
	}
	if outcome.Status != sdk.TerminalStatusCompleted {
		t.Fatalf("expected completed status, got %s", outcome.Status)
	}
	if outcome.Text != "done" {
		t.Fatalf("expected text done, got %q", outcome.Text)
	}
	if outcome.StopReason != "stop" {
		t.Fatalf("expected stop reason stop, got %q", outcome.StopReason)
	}
}

func TestDecodeTerminalOutcomeAborted(t *testing.T) {
	raw := json.RawMessage(`{
		"type":"agent_end",
		"messages":[
			{"role":"assistant","content":[{"type":"text","text":"partial"}],"stopReason":"aborted"}
		]
	}`)

	outcome, err := sdk.DecodeTerminalOutcome(raw)
	if err != nil {
		t.Fatalf("DecodeTerminalOutcome returned error: %v", err)
	}
	if outcome.Status != sdk.TerminalStatusAborted {
		t.Fatalf("expected aborted status, got %s", outcome.Status)
	}
}

func TestDecodeTerminalOutcomeFailed(t *testing.T) {
	raw := json.RawMessage(`{
		"type":"agent_end",
		"messages":[
			{"role":"assistant","content":[{"type":"text","text":""}],"stopReason":"error","errorMessage":"provider failed"}
		]
	}`)

	outcome, err := sdk.DecodeTerminalOutcome(raw)
	if err != nil {
		t.Fatalf("DecodeTerminalOutcome returned error: %v", err)
	}
	if outcome.Status != sdk.TerminalStatusFailed {
		t.Fatalf("expected failed status, got %s", outcome.Status)
	}
	if outcome.ErrorMessage != "provider failed" {
		t.Fatalf("expected error message, got %q", outcome.ErrorMessage)
	}
}
