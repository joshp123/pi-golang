package testsupport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

const (
	commandPrompt      = "prompt"
	commandAbort       = "abort"
	commandGetState    = "get_state"
	commandNewSession  = "new_session"
	commandCompact     = "compact"
	commandGetCommands = "get_commands"

	eventTypeResponse            = "response"
	eventTypeAgentEnd            = "agent_end"
	eventTypeMessageUpdate       = "message_update"
	eventTypeAutoCompactionStart = "auto_compaction_start"
	eventTypeAutoCompactionEnd   = "auto_compaction_end"
	eventTypeAutoRetryStart      = "auto_retry_start"
	eventTypeAutoRetryEnd        = "auto_retry_end"
)

func RunScenario(scenario string, processArgs []string, stdin io.Reader, stdout io.Writer) error {
	scanner := bufio.NewScanner(stdin)
	writer := bufio.NewWriter(stdout)
	defer writer.Flush()

	abortRun := abortRunState{}
	runCancelAbort := runCancelAbortState{}
	skillPaths := collectFlagValues(processArgs, "--skill")

	for scanner.Scan() {
		var command map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &command); err != nil {
			return err
		}

		commandType, _ := command["type"].(string)
		requestID, _ := command["id"].(string)

		if commandType == commandGetCommands {
			commands := make([]map[string]any, 0, len(skillPaths))
			for index, path := range skillPaths {
				commands = append(commands, map[string]any{
					"name":        fmt.Sprintf("skill:test-%d", index+1),
					"description": "test skill",
					"source":      "skill",
					"location":    "path",
					"path":        path,
				})
			}
			if scenario == "skills_unexpected" {
				commands = append(commands, map[string]any{
					"name":        "skill:ambient",
					"description": "ambient skill",
					"source":      "skill",
					"location":    "user",
					"path":        "/unexpected/ambient/SKILL.md",
				})
			}
			if err := writeResponse(writer, requestID, commandType, true, map[string]any{"commands": commands}, ""); err != nil {
				return err
			}
			continue
		}

		switch scenario {
		case "die_on_prompt":
			if commandType == commandPrompt {
				return nil
			}
			if err := writeResponse(writer, requestID, commandType, true, map[string]any{}, ""); err != nil {
				return err
			}
		case "happy", "skills_unexpected":
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
		case "run_detailed_signals":
			if err := handleRunDetailedSignalsScenario(writer, requestID, commandType); err != nil {
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

func collectFlagValues(args []string, flag string) []string {
	values := make([]string, 0)
	for index := 0; index < len(args)-1; index++ {
		if args[index] != flag {
			continue
		}
		values = append(values, args[index+1])
		index++
	}
	return values
}
