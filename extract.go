package pi

import (
	"encoding/json"
	"fmt"
	"strings"
)

func debugExtract(format string, args ...any) {
	if Debug {
		debugf("extract: "+format, args...)
	}
}

type contentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

func decodeRPCResponse(raw json.RawMessage) (rpcResponse, error) {
	var response rpcResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return rpcResponse{}, err
	}
	if err := requireEnvelopeType("response", response.Type, eventTypeResponse); err != nil {
		return rpcResponse{}, err
	}
	if strings.TrimSpace(response.Command) == "" {
		return rpcResponse{}, fmt.Errorf("%w: response missing command", ErrProtocolViolation)
	}
	return response, nil
}

func DecodeAgentEnd(raw json.RawMessage) (AgentEndEvent, error) {
	var payload struct {
		Type     string         `json:"type"`
		Messages []AgentMessage `json:"messages"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return AgentEndEvent{}, err
	}
	if err := requireEnvelopeType("event", payload.Type, EventTypeAgentEnd); err != nil {
		return AgentEndEvent{}, err
	}
	return AgentEndEvent{Messages: payload.Messages}, nil
}

func DecodeMessageUpdate(raw json.RawMessage) (MessageUpdateEvent, error) {
	var payload struct {
		Type                  string          `json:"type"`
		Message               AgentMessage    `json:"message"`
		AssistantMessageEvent json.RawMessage `json:"assistantMessageEvent"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return MessageUpdateEvent{}, err
	}
	if err := requireEnvelopeType("event", payload.Type, EventTypeMessageUpdate); err != nil {
		return MessageUpdateEvent{}, err
	}

	var delta AssistantMessageDelta
	if len(payload.AssistantMessageEvent) > 0 && string(payload.AssistantMessageEvent) != "null" {
		if err := json.Unmarshal(payload.AssistantMessageEvent, &delta); err != nil {
			return MessageUpdateEvent{}, err
		}
		delta.Raw = append([]byte(nil), payload.AssistantMessageEvent...)
	}

	return MessageUpdateEvent{
		Message:               payload.Message,
		AssistantMessageEvent: delta,
	}, nil
}

func DecodeAutoCompactionEnd(raw json.RawMessage) (AutoCompactionEndEvent, error) {
	var payload struct {
		Type         string         `json:"type"`
		Result       *CompactResult `json:"result"`
		Aborted      bool           `json:"aborted"`
		WillRetry    bool           `json:"willRetry"`
		ErrorMessage string         `json:"errorMessage"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return AutoCompactionEndEvent{}, err
	}
	if err := requireEnvelopeType("event", payload.Type, EventTypeAutoCompactionEnd); err != nil {
		return AutoCompactionEndEvent{}, err
	}
	return AutoCompactionEndEvent{
		Result:       payload.Result,
		Aborted:      payload.Aborted,
		WillRetry:    payload.WillRetry,
		ErrorMessage: payload.ErrorMessage,
	}, nil
}

func requireEnvelopeType(kind string, actual string, expected string) error {
	if strings.TrimSpace(actual) == "" {
		return fmt.Errorf("%w: %s missing type", ErrProtocolViolation, kind)
	}
	if actual != expected {
		return fmt.Errorf("unexpected %s type %q", kind, actual)
	}
	return nil
}

func extractRunResult(event Event) (RunResult, error) {
	debugExtract("parsing agent_end payload (%d bytes)", len(event.Raw))

	outcome, err := DecodeTerminalOutcome(event.Raw)
	if err != nil {
		debugExtract("decode error: %v", err)
		return RunResult{}, err
	}
	debugExtract("terminal status=%s", outcome.Status)
	return RunResult{Text: outcome.Text, Usage: outcome.Usage}, nil
}

func extractAssistantText(content json.RawMessage) (string, error) {
	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" || trimmed == "null" {
		return "", nil
	}
	if strings.HasPrefix(trimmed, "\"") {
		var text string
		if err := json.Unmarshal(content, &text); err != nil {
			return "", err
		}
		return text, nil
	}

	var blocks []contentBlock
	if err := json.Unmarshal(content, &blocks); err != nil {
		return "", err
	}

	var builder strings.Builder
	for _, block := range blocks {
		if block.Type == "text" {
			builder.WriteString(block.Text)
		}
	}
	return builder.String(), nil
}
