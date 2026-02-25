package testsupport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"
)

func handleHappyScenario(writer *bufio.Writer, requestID string, commandType string, command map[string]any) error {
	switch commandType {
	case commandGetState:
		return writeResponse(writer, requestID, commandType, true, map[string]any{
			"sessionId":             "session-123",
			"sessionFile":           "/tmp/session-123.jsonl",
			"autoCompactionEnabled": true,
			"model": map[string]any{
				"id":            "claude-opus-4-5",
				"provider":      "anthropic",
				"contextWindow": 200000,
				"maxTokens":     8192,
			},
		}, "")
	case commandNewSession:
		parent, _ := command["parentSession"].(string)
		cancelled := parent == "cancel-parent"
		return writeResponse(writer, requestID, commandType, true, map[string]any{"cancelled": cancelled}, "")
	case commandCompact:
		customInstructions, _ := command["customInstructions"].(string)
		if customInstructions == "force-error" {
			return writeResponse(writer, requestID, commandType, false, nil, "compact failed")
		}
		if err := writeResponse(writer, requestID, commandType, true, map[string]any{
			"summary":          "compacted",
			"firstKeptEntryId": "entry-1",
			"tokensBefore":     12345,
			"details":          map[string]any{"mode": "manual"},
		}, ""); err != nil {
			return err
		}
	case commandPrompt:
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		if err := writeEvent(writer, map[string]any{
			"type": eventTypeMessageUpdate,
			"message": map[string]any{
				"role":    "assistant",
				"content": []map[string]any{{"type": "text", "text": "hello"}},
			},
			"assistantMessageEvent": map[string]any{
				"type":         "text_delta",
				"contentIndex": 0,
				"delta":        "hello",
			},
		}); err != nil {
			return err
		}
		if err := writeEvent(writer, map[string]any{
			"type": eventTypeAgentEnd,
			"messages": []map[string]any{
				{"role": "user", "content": "hello"},
				{
					"role": "assistant",
					"content": []map[string]any{
						{"type": "text", "text": "hello from helper"},
					},
					"usage": map[string]any{"input": 10, "output": 5, "cacheRead": 0, "cacheWrite": 0},
				},
			},
		}); err != nil {
			return err
		}
	case commandAbort:
		return writeResponse(writer, requestID, commandType, true, nil, "")
	default:
		return writeResponse(writer, requestID, commandType, false, nil, "unknown command")
	}
	return nil
}

func handlePromptAsyncErrorScenario(writer *bufio.Writer, requestID string, commandType string) error {
	switch commandType {
	case commandPrompt:
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		return writeResponse(writer, requestID, commandType, false, nil, "streamingBehavior is required while streaming")
	default:
		return writeResponse(writer, requestID, commandType, true, map[string]any{}, "")
	}
}

func handleFloodBeforeResponseScenario(writer *bufio.Writer, requestID string, commandType string) error {
	switch commandType {
	case commandGetState:
		for index := 0; index < 128; index++ {
			if err := writeEvent(writer, map[string]any{
				"type": eventTypeMessageUpdate,
				"message": map[string]any{
					"role":    "assistant",
					"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("chunk-%d", index)}},
				},
				"assistantMessageEvent": map[string]any{"type": "text_delta", "delta": "chunk"},
			}); err != nil {
				return err
			}
		}
		return writeResponse(writer, requestID, commandType, true, map[string]any{
			"sessionId":   "session-123",
			"sessionFile": "/tmp/session-123.jsonl",
			"model": map[string]any{
				"id":            "claude-opus-4-5",
				"provider":      "anthropic",
				"contextWindow": 200000,
				"maxTokens":     8192,
			},
		}, "")
	default:
		return writeResponse(writer, requestID, commandType, true, map[string]any{}, "")
	}
}

func handleSlowRunScenario(writer *bufio.Writer, requestID string, commandType string) error {
	switch commandType {
	case commandPrompt:
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond)
		return writeEvent(writer, map[string]any{
			"type": eventTypeAgentEnd,
			"messages": []map[string]any{
				{
					"role":    "assistant",
					"content": []map[string]any{{"type": "text", "text": "done"}},
				},
			},
		})
	default:
		return writeResponse(writer, requestID, commandType, true, map[string]any{}, "")
	}
}

type abortRunState struct {
	promptSeen   bool
	abortSeen    bool
	agentEndSent bool
}

func handleAbortRunScenario(writer *bufio.Writer, state *abortRunState, requestID string, commandType string) error {
	switch commandType {
	case commandPrompt:
		state.promptSeen = true
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		return maybeWriteAbortRunAgentEnd(writer, state)
	case commandAbort:
		state.abortSeen = true
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		return maybeWriteAbortRunAgentEnd(writer, state)
	default:
		return writeResponse(writer, requestID, commandType, true, map[string]any{}, "")
	}
}

func maybeWriteAbortRunAgentEnd(writer *bufio.Writer, state *abortRunState) error {
	if state.agentEndSent || !state.promptSeen || !state.abortSeen {
		return nil
	}
	state.agentEndSent = true
	return writeEvent(writer, map[string]any{
		"type": eventTypeAgentEnd,
		"messages": []map[string]any{
			{
				"role":       "assistant",
				"content":    []map[string]any{{"type": "text", "text": "aborted by helper"}},
				"stopReason": "aborted",
			},
		},
	})
}

type runCancelAbortState struct {
	promptSeen bool
	abortSeen  bool
}

func handleRunCancelAbortScenario(writer *bufio.Writer, state *runCancelAbortState, requestID string, commandType string) error {
	switch commandType {
	case commandPrompt:
		state.promptSeen = true
		return writeResponse(writer, requestID, commandType, true, nil, "")
	case commandAbort:
		state.abortSeen = true
		return writeResponse(writer, requestID, commandType, true, nil, "")
	case commandGetState:
		if !state.promptSeen || !state.abortSeen {
			return writeResponse(writer, requestID, commandType, false, nil, "abort not called")
		}
		return writeResponse(writer, requestID, commandType, true, map[string]any{
			"sessionId": "abort-observed",
			"model": map[string]any{
				"id":            "claude-opus-4-5",
				"provider":      "anthropic",
				"contextWindow": 200000,
				"maxTokens":     8192,
			},
		}, "")
	default:
		return writeResponse(writer, requestID, commandType, true, map[string]any{}, "")
	}
}

func handleRunDetailedSignalsScenario(writer *bufio.Writer, requestID string, commandType string) error {
	switch commandType {
	case commandPrompt:
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		if err := writeEvent(writer, map[string]any{"type": eventTypeAutoCompactionStart, "reason": "overflow"}); err != nil {
			return err
		}
		if err := writeEvent(writer, map[string]any{"type": eventTypeAutoCompactionEnd, "result": map[string]any{"summary": "compacted", "firstKeptEntryId": "entry-1", "tokensBefore": 120000}, "aborted": false, "willRetry": true}); err != nil {
			return err
		}
		if err := writeEvent(writer, map[string]any{"type": eventTypeAutoRetryStart, "attempt": 1, "maxAttempts": 3, "delayMs": 10, "errorMessage": "overflow"}); err != nil {
			return err
		}
		if err := writeEvent(writer, map[string]any{"type": eventTypeAutoRetryEnd, "success": true, "attempt": 1}); err != nil {
			return err
		}
		return writeEvent(writer, map[string]any{
			"type": eventTypeAgentEnd,
			"messages": []map[string]any{
				{
					"role":    "assistant",
					"content": []map[string]any{{"type": "text", "text": "after compaction"}},
					"usage":   map[string]any{"input": 10, "output": 5, "cacheRead": 0, "cacheWrite": 0},
				},
			},
		})
	default:
		return writeResponse(writer, requestID, commandType, true, map[string]any{}, "")
	}
}

func writeResponse(writer *bufio.Writer, id string, command string, success bool, data any, errText string) error {
	payload := map[string]any{
		"type":    eventTypeResponse,
		"id":      id,
		"command": command,
		"success": success,
	}
	if data != nil {
		payload["data"] = data
	}
	if errText != "" {
		payload["error"] = errText
	}
	return writeEvent(writer, payload)
}

func writeEvent(writer *bufio.Writer, payload map[string]any) error {
	line, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := writer.Write(append(line, '\n')); err != nil {
		return err
	}
	return writer.Flush()
}
