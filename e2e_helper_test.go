package pi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupFakePI(t *testing.T, scenario string) {
	t.Helper()

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable failed: %v", err)
	}

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "pi")
	script := fmt.Sprintf("#!/usr/bin/env bash\nset -euo pipefail\nGO_WANT_PI_HELPER=1 exec %q -test.run '^TestHelperProcess$' -- %q \"$@\"\n", exe, scenario)
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake pi: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(t.TempDir(), "pi-agent"))
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_PI_HELPER") != "1" {
		return
	}

	scenario := "happy"
	for index, arg := range os.Args {
		if arg == "--" && index+1 < len(os.Args) {
			scenario = os.Args[index+1]
			break
		}
	}

	if err := runHelperScenario(scenario); err != nil {
		fmt.Fprintf(os.Stderr, "helper scenario failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runHelperScenario(scenario string) error {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	abortRun := helperAbortRunState{}
	runCancelAbort := helperRunCancelAbortState{}

	for scanner.Scan() {
		var command map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &command); err != nil {
			return err
		}

		commandType, _ := command["type"].(string)
		requestID, _ := command["id"].(string)

		switch scenario {
		case "die_on_prompt":
			if commandType == rpcCommandPrompt {
				return nil
			}
			if err := writeResponse(writer, requestID, commandType, true, map[string]any{}, ""); err != nil {
				return err
			}
		case "happy":
			if err := handleHappyScenario(writer, requestID, commandType, command); err != nil {
				return err
			}
		case "prompt_async_error":
			if err := handlePromptAsyncErrorScenario(writer, requestID, commandType); err != nil {
				return err
			}
		case "flood_before_response":
			if err := handleFloodBeforeResponseScenario(writer, requestID, commandType); err != nil {
				return err
			}
		case "slow_run":
			if err := handleSlowRunScenario(writer, requestID, commandType); err != nil {
				return err
			}
		case "abort_run":
			if err := handleAbortRunScenario(writer, &abortRun, requestID, commandType); err != nil {
				return err
			}
		case "run_ctx_cancel_aborts":
			if err := handleRunCancelAbortScenario(writer, &runCancelAbort, requestID, commandType); err != nil {
				return err
			}
		case "never_respond":
			continue
		default:
			return fmt.Errorf("unknown scenario %q", scenario)
		}
	}

	return scanner.Err()
}

func handleHappyScenario(writer *bufio.Writer, requestID string, commandType string, command map[string]any) error {
	switch commandType {
	case rpcCommandGetState:
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
	case rpcCommandNewSession:
		parent, _ := command["parentSession"].(string)
		cancelled := parent == "cancel-parent"
		return writeResponse(writer, requestID, commandType, true, map[string]any{"cancelled": cancelled}, "")
	case rpcCommandCompact:
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
	case rpcCommandPrompt:
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		if err := writeEvent(writer, map[string]any{
			"type": EventTypeMessageUpdate,
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
			"type": EventTypeAgentEnd,
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
	case rpcCommandAbort:
		return writeResponse(writer, requestID, commandType, true, nil, "")
	default:
		return writeResponse(writer, requestID, commandType, false, nil, "unknown command")
	}
	return nil
}

func handlePromptAsyncErrorScenario(writer *bufio.Writer, requestID string, commandType string) error {
	switch commandType {
	case rpcCommandPrompt:
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
	case rpcCommandGetState:
		for index := 0; index < 128; index++ {
			if err := writeEvent(writer, map[string]any{
				"type": EventTypeMessageUpdate,
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
	case rpcCommandPrompt:
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond)
		return writeEvent(writer, map[string]any{
			"type": EventTypeAgentEnd,
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

type helperAbortRunState struct {
	promptSeen   bool
	abortSeen    bool
	agentEndSent bool
}

func handleAbortRunScenario(writer *bufio.Writer, state *helperAbortRunState, requestID string, commandType string) error {
	switch commandType {
	case rpcCommandPrompt:
		state.promptSeen = true
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		return maybeWriteAbortRunAgentEnd(writer, state)
	case rpcCommandAbort:
		state.abortSeen = true
		if err := writeResponse(writer, requestID, commandType, true, nil, ""); err != nil {
			return err
		}
		return maybeWriteAbortRunAgentEnd(writer, state)
	default:
		return writeResponse(writer, requestID, commandType, true, map[string]any{}, "")
	}
}

func maybeWriteAbortRunAgentEnd(writer *bufio.Writer, state *helperAbortRunState) error {
	if state.agentEndSent || !state.promptSeen || !state.abortSeen {
		return nil
	}
	state.agentEndSent = true
	return writeEvent(writer, map[string]any{
		"type": EventTypeAgentEnd,
		"messages": []map[string]any{
			{
				"role":       "assistant",
				"content":    []map[string]any{{"type": "text", "text": "aborted by helper"}},
				"stopReason": "aborted",
			},
		},
	})
}

type helperRunCancelAbortState struct {
	promptSeen bool
	abortSeen  bool
}

func handleRunCancelAbortScenario(writer *bufio.Writer, state *helperRunCancelAbortState, requestID string, commandType string) error {
	switch commandType {
	case rpcCommandPrompt:
		state.promptSeen = true
		return writeResponse(writer, requestID, commandType, true, nil, "")
	case rpcCommandAbort:
		state.abortSeen = true
		return writeResponse(writer, requestID, commandType, true, nil, "")
	case rpcCommandGetState:
		if !state.promptSeen || !state.abortSeen {
			return writeResponse(writer, requestID, commandType, false, nil, "abort not called")
		}
		return writeResponse(writer, requestID, commandType, true, map[string]any{
			"sessionId": "abort-observed",
		}, "")
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

func readEventOrFail(t *testing.T, events <-chan Event) Event {
	t.Helper()
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed")
		}
		return event
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
		return Event{}
	}
}
