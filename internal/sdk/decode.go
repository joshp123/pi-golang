package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func debugExtract(format string, args ...any) {
	if debugEnabledProvider() {
		debugf("extract: "+format, args...)
	}
}

type contentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
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

func DecodeAutoCompactionStart(raw json.RawMessage) (AutoCompactionStartEvent, error) {
	var payload struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return AutoCompactionStartEvent{}, err
	}
	if err := requireEnvelopeType("event", payload.Type, EventTypeAutoCompactionStart); err != nil {
		return AutoCompactionStartEvent{}, err
	}
	return AutoCompactionStartEvent{Reason: payload.Reason}, nil
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

func DecodeAutoRetryStart(raw json.RawMessage) (AutoRetryStartEvent, error) {
	var payload struct {
		Type         string `json:"type"`
		Attempt      int    `json:"attempt"`
		MaxAttempts  int    `json:"maxAttempts"`
		DelayMS      int    `json:"delayMs"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return AutoRetryStartEvent{}, err
	}
	if err := requireEnvelopeType("event", payload.Type, EventTypeAutoRetryStart); err != nil {
		return AutoRetryStartEvent{}, err
	}
	return AutoRetryStartEvent{
		Attempt:      payload.Attempt,
		MaxAttempts:  payload.MaxAttempts,
		DelayMS:      payload.DelayMS,
		ErrorMessage: payload.ErrorMessage,
	}, nil
}

func DecodeAutoRetryEnd(raw json.RawMessage) (AutoRetryEndEvent, error) {
	var payload struct {
		Type       string `json:"type"`
		Success    bool   `json:"success"`
		Attempt    int    `json:"attempt"`
		FinalError string `json:"finalError"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return AutoRetryEndEvent{}, err
	}
	if err := requireEnvelopeType("event", payload.Type, EventTypeAutoRetryEnd); err != nil {
		return AutoRetryEndEvent{}, err
	}
	return AutoRetryEndEvent{
		Success:    payload.Success,
		Attempt:    payload.Attempt,
		FinalError: payload.FinalError,
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

// DecodeTerminalOutcome converts an agent_end payload into one canonical terminal shape.
func DecodeTerminalOutcome(raw json.RawMessage) (TerminalOutcome, error) {
	agentEnd, err := DecodeAgentEnd(raw)
	if err != nil {
		return TerminalOutcome{}, err
	}

	for index := len(agentEnd.Messages) - 1; index >= 0; index-- {
		message := agentEnd.Messages[index]
		if message.Role != "assistant" {
			continue
		}
		text, err := extractAssistantText(message.Content)
		if err != nil {
			return TerminalOutcome{}, err
		}
		return TerminalOutcome{
			Status:         terminalStatus(message.StopReason, message.ErrorMessage),
			Text:           text,
			StopReason:     strings.TrimSpace(message.StopReason),
			TerminalReason: terminalReasonFromMessage(message),
			ErrorMessage:   strings.TrimSpace(message.ErrorMessage),
			Usage:          message.Usage,
		}, nil
	}

	return TerminalOutcome{}, errors.New("assistant message not found in agent_end")
}

func terminalReasonFromMessage(message AgentMessage) TerminalReason {
	candidates := []TerminalReason{
		message.TerminalReason,
		message.TerminalReasonAlt,
	}

	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(string(candidate))
		if trimmed != "" {
			return TerminalReason(trimmed)
		}
	}
	return ""
}

func terminalStatus(stopReason string, errorMessage string) TerminalStatus {
	reason := strings.ToLower(strings.TrimSpace(stopReason))
	errText := strings.TrimSpace(errorMessage)

	if reason == "aborted" {
		return TerminalStatusAborted
	}
	if errText != "" {
		return TerminalStatusFailed
	}
	switch reason {
	case "error", "failed":
		return TerminalStatusFailed
	default:
		return TerminalStatusCompleted
	}
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
