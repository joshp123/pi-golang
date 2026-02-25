package pi

import (
	"encoding/json"
	"errors"
	"strings"
)

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
			Status:       terminalStatus(message.StopReason, message.ErrorMessage),
			Text:         text,
			StopReason:   strings.TrimSpace(message.StopReason),
			ErrorMessage: strings.TrimSpace(message.ErrorMessage),
			Usage:        message.Usage,
		}, nil
	}

	return TerminalOutcome{}, errors.New("assistant message not found in agent_end")
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
