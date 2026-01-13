package pi

import (
	"encoding/json"
	"errors"
	"strings"
)

type agentEndPayload struct {
	Messages []agentMessage `json:"messages"`
}

type agentMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Usage   *Usage          `json:"usage,omitempty"`
}

type contentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

func extractRunResult(event Event) (RunResult, error) {
	var payload agentEndPayload
	if err := json.Unmarshal(event.Raw, &payload); err != nil {
		return RunResult{}, err
	}

	for index := len(payload.Messages) - 1; index >= 0; index-- {
		message := payload.Messages[index]
		if message.Role != "assistant" {
			continue
		}
		text, err := extractAssistantText(message.Content)
		if err != nil {
			return RunResult{}, err
		}
		return RunResult{Text: text, Usage: message.Usage}, nil
	}

	return RunResult{}, errors.New("assistant message not found in agent_end")
}

func extractAssistantText(content json.RawMessage) (string, error) {
	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
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

	var parts []string
	for _, block := range blocks {
		if block.Type == "text" && block.Text != "" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, ""), nil
}
