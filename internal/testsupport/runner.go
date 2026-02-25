package testsupport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

const (
	commandPrompt     = "prompt"
	commandAbort      = "abort"
	commandGetState   = "get_state"
	commandNewSession = "new_session"
	commandCompact    = "compact"

	eventTypeResponse      = "response"
	eventTypeAgentEnd      = "agent_end"
	eventTypeMessageUpdate = "message_update"
)

func RunScenario(scenario string, stdin io.Reader, stdout io.Writer) error {
	scanner := bufio.NewScanner(stdin)
	writer := bufio.NewWriter(stdout)
	defer writer.Flush()

	abortRun := abortRunState{}
	runCancelAbort := runCancelAbortState{}

	for scanner.Scan() {
		var command map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &command); err != nil {
			return err
		}

		commandType, _ := command["type"].(string)
		requestID, _ := command["id"].(string)

		switch scenario {
		case "die_on_prompt":
			if commandType == commandPrompt {
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
