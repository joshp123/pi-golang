package sdk

import "testing"

func TestExtractRunResult(t *testing.T) {
	eventJSON := `{"type":"agent_end","messages":[{"role":"user","content":"hi"},{"role":"assistant","content":[{"type":"text","text":"hello"},{"type":"thinking","thinking":"..."}],"usage":{"input":10,"output":5,"cacheRead":1,"cacheWrite":2,"cost":{"input":0.1,"output":0.2,"cacheRead":0.01,"cacheWrite":0.02,"total":0.33}}}]}`
	result, err := extractRunResult(Event{Type: EventTypeAgentEnd, Raw: []byte(eventJSON)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "hello" {
		t.Fatalf("expected text hello, got %q", result.Text)
	}
	if result.Usage == nil || result.Usage.Input != 10 {
		t.Fatalf("expected usage input 10")
	}
}

func TestExtractAssistantTextString(t *testing.T) {
	result, err := extractAssistantText([]byte(`"plain"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "plain" {
		t.Fatalf("expected plain, got %q", result)
	}
}
