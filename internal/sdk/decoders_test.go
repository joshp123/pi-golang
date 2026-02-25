package sdk

import (
	"encoding/json"
	"testing"

	"github.com/joshp123/pi-golang/internal/rpc"
)

func TestDecodeAgentEnd(t *testing.T) {
	raw := json.RawMessage(`{"type":"agent_end","messages":[{"role":"assistant","content":[{"type":"text","text":"hello"}],"usage":{"input":1,"output":2}}]}`)
	event, err := DecodeAgentEnd(raw)
	if err != nil {
		t.Fatalf("DecodeAgentEnd returned error: %v", err)
	}
	if len(event.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(event.Messages))
	}
	if event.Messages[0].Role != "assistant" {
		t.Fatalf("unexpected role: %s", event.Messages[0].Role)
	}
}

func TestDecodeMessageUpdate(t *testing.T) {
	raw := json.RawMessage(`{
		"type":"message_update",
		"message":{"role":"assistant","content":[{"type":"text","text":"partial"}]},
		"assistantMessageEvent":{"type":"text_delta","contentIndex":0,"delta":"partial"}
	}`)
	event, err := DecodeMessageUpdate(raw)
	if err != nil {
		t.Fatalf("DecodeMessageUpdate returned error: %v", err)
	}
	if event.AssistantMessageEvent.Type != "text_delta" {
		t.Fatalf("expected text_delta, got %s", event.AssistantMessageEvent.Type)
	}
	if event.AssistantMessageEvent.Delta != "partial" {
		t.Fatalf("expected delta partial, got %q", event.AssistantMessageEvent.Delta)
	}
}

func TestDecodeRPCResponse(t *testing.T) {
	raw := json.RawMessage(`{"type":"response","id":"req-1","command":"prompt","success":false,"error":"bad"}`)
	response, err := decodeRPCResponse(raw)
	if err != nil {
		t.Fatalf("decodeRPCResponse returned error: %v", err)
	}
	if response.ID != "req-1" || response.Command != rpc.CommandPrompt {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Success {
		t.Fatal("expected success=false")
	}
}

func TestDecodeRPCResponseRequiresType(t *testing.T) {
	raw := json.RawMessage(`{"id":"req-1","command":"prompt","success":true}`)
	_, err := decodeRPCResponse(raw)
	if err == nil {
		t.Fatal("expected missing type error")
	}
}

func TestDecodeAgentEndRequiresType(t *testing.T) {
	raw := json.RawMessage(`{"messages":[]}`)
	_, err := DecodeAgentEnd(raw)
	if err == nil {
		t.Fatal("expected missing type error")
	}
}

func TestDecodeMessageUpdateRequiresType(t *testing.T) {
	raw := json.RawMessage(`{"message":{"role":"assistant","content":[]}}`)
	_, err := DecodeMessageUpdate(raw)
	if err == nil {
		t.Fatal("expected missing type error")
	}
}

func TestDecodeAutoCompactionEndRequiresType(t *testing.T) {
	raw := json.RawMessage(`{"aborted":false,"willRetry":false}`)
	_, err := DecodeAutoCompactionEnd(raw)
	if err == nil {
		t.Fatal("expected missing type error")
	}
}

func TestDecodeAutoCompactionEndVariants(t *testing.T) {
	tests := []struct {
		name        string
		raw         json.RawMessage
		wantAborted bool
		wantRetry   bool
		wantError   string
		wantResult  bool
	}{
		{
			name: "success",
			raw: json.RawMessage(`{
				"type":"auto_compaction_end",
				"result":{"summary":"done","firstKeptEntryId":"abc","tokensBefore":123,"details":{}},
				"aborted":false,
				"willRetry":true
			}`),
			wantRetry:  true,
			wantResult: true,
		},
		{
			name: "aborted",
			raw: json.RawMessage(`{
				"type":"auto_compaction_end",
				"result":null,
				"aborted":true,
				"willRetry":false
			}`),
			wantAborted: true,
		},
		{
			name: "error",
			raw: json.RawMessage(`{
				"type":"auto_compaction_end",
				"result":null,
				"aborted":false,
				"willRetry":false,
				"errorMessage":"quota exceeded"
			}`),
			wantError: "quota exceeded",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event, err := DecodeAutoCompactionEnd(tc.raw)
			if err != nil {
				t.Fatalf("DecodeAutoCompactionEnd returned error: %v", err)
			}
			if event.Aborted != tc.wantAborted {
				t.Fatalf("unexpected aborted: got=%v want=%v", event.Aborted, tc.wantAborted)
			}
			if event.WillRetry != tc.wantRetry {
				t.Fatalf("unexpected willRetry: got=%v want=%v", event.WillRetry, tc.wantRetry)
			}
			if event.ErrorMessage != tc.wantError {
				t.Fatalf("unexpected error message: got=%q want=%q", event.ErrorMessage, tc.wantError)
			}
			if (event.Result != nil) != tc.wantResult {
				t.Fatalf("unexpected result presence: got=%v want=%v", event.Result != nil, tc.wantResult)
			}
		})
	}
}

func TestDecodeAutoCompactionStart(t *testing.T) {
	raw := json.RawMessage(`{"type":"auto_compaction_start","reason":"overflow"}`)
	event, err := DecodeAutoCompactionStart(raw)
	if err != nil {
		t.Fatalf("DecodeAutoCompactionStart returned error: %v", err)
	}
	if event.Reason != "overflow" {
		t.Fatalf("expected reason overflow, got %q", event.Reason)
	}
}

func TestDecodeAutoRetryStart(t *testing.T) {
	raw := json.RawMessage(`{"type":"auto_retry_start","attempt":1,"maxAttempts":3,"delayMs":2000,"errorMessage":"overloaded"}`)
	event, err := DecodeAutoRetryStart(raw)
	if err != nil {
		t.Fatalf("DecodeAutoRetryStart returned error: %v", err)
	}
	if event.Attempt != 1 || event.MaxAttempts != 3 || event.DelayMS != 2000 {
		t.Fatalf("unexpected auto_retry_start payload: %+v", event)
	}
}

func TestDecodeAutoRetryEnd(t *testing.T) {
	raw := json.RawMessage(`{"type":"auto_retry_end","success":false,"attempt":3,"finalError":"quota"}`)
	event, err := DecodeAutoRetryEnd(raw)
	if err != nil {
		t.Fatalf("DecodeAutoRetryEnd returned error: %v", err)
	}
	if event.Success || event.Attempt != 3 || event.FinalError != "quota" {
		t.Fatalf("unexpected auto_retry_end payload: %+v", event)
	}
}
