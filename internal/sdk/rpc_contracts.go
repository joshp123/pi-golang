package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/joshp123/pi-golang/internal/rpc"
)

// Thin RPC command builders/decoders. Keep these 1:1 with upstream rpc.md.

func promptCommand(request PromptRequest) (rpc.Command, error) {
	if err := validatePromptRequest(request, true); err != nil {
		return nil, err
	}
	command := rpc.Command{
		"type":    rpc.CommandPrompt,
		"message": request.Message,
	}
	if request.StreamingBehavior != "" {
		command["streamingBehavior"] = string(request.StreamingBehavior)
	}
	if len(request.Images) > 0 {
		command["images"] = encodeImages(request.Images)
	}
	return command, nil
}

func steerCommand(request PromptRequest) (rpc.Command, error) {
	return queuedMessageCommand(rpc.CommandSteer, request)
}

func followUpCommand(request PromptRequest) (rpc.Command, error) {
	return queuedMessageCommand(rpc.CommandFollowUp, request)
}

func queuedMessageCommand(commandType string, request PromptRequest) (rpc.Command, error) {
	if err := validatePromptRequest(request, false); err != nil {
		return nil, err
	}
	command := rpc.Command{
		"type":    commandType,
		"message": request.Message,
	}
	if len(request.Images) > 0 {
		command["images"] = encodeImages(request.Images)
	}
	return command, nil
}

func abortCommand() rpc.Command {
	return rpc.Command{"type": rpc.CommandAbort}
}

func getStateCommand() rpc.Command {
	return rpc.Command{"type": rpc.CommandGetState}
}

func newSessionCommand(parentSession string) rpc.Command {
	command := rpc.Command{"type": rpc.CommandNewSession}
	if strings.TrimSpace(parentSession) != "" {
		command["parentSession"] = strings.TrimSpace(parentSession)
	}
	return command
}

func compactCommand(customInstructions string) rpc.Command {
	command := rpc.Command{"type": rpc.CommandCompact}
	if strings.TrimSpace(customInstructions) != "" {
		command["customInstructions"] = strings.TrimSpace(customInstructions)
	}
	return command
}

func exportHTMLCommand(outputPath string) rpc.Command {
	command := rpc.Command{"type": rpc.CommandExportHTML}
	if strings.TrimSpace(outputPath) != "" {
		command["outputPath"] = outputPath
	}
	return command
}

func getCommandsCommand() rpc.Command {
	return rpc.Command{"type": rpc.CommandGetCommands}
}

type slashCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Location    string `json:"location"`
	Path        string `json:"path"`
}

func decodeCommands(data json.RawMessage) ([]slashCommand, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, fmt.Errorf("%w: get_commands missing response data", ErrProtocolViolation)
	}
	var payload struct {
		Commands []slashCommand `json:"commands"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload.Commands, nil
}

func decodeSessionState(data json.RawMessage) (SessionState, error) {
	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return SessionState{}, err
	}
	if state.SessionID == "" {
		return SessionState{}, fmt.Errorf("%w: get_state missing sessionId", ErrProtocolViolation)
	}
	if state.ContextWindow == 0 && state.Model != nil {
		state.ContextWindow = state.Model.ContextWindow
	}
	if state.ContextWindow <= 0 {
		return SessionState{}, fmt.Errorf("%w: get_state missing context window", ErrProtocolViolation)
	}
	return state, nil
}

func decodeNewSessionCancelled(data json.RawMessage) (bool, error) {
	if len(data) == 0 || string(data) == "null" {
		return false, fmt.Errorf("%w: new_session missing response data", ErrProtocolViolation)
	}

	var payload struct {
		Cancelled bool `json:"cancelled"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return false, err
	}
	return payload.Cancelled, nil
}

func decodeCompactResult(data json.RawMessage) (CompactResult, error) {
	var result CompactResult
	if err := json.Unmarshal(data, &result); err != nil {
		return CompactResult{}, err
	}
	return result, nil
}

func decodeExportPath(data json.RawMessage) (string, error) {
	var payload struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.Path) == "" {
		return "", errors.New("export_html returned empty path")
	}
	return payload.Path, nil
}

func decodeRPCResponse(raw json.RawMessage) (rpc.Response, error) {
	var response rpc.Response
	if err := json.Unmarshal(raw, &response); err != nil {
		return rpc.Response{}, err
	}
	if err := requireEnvelopeType("response", response.Type, rpc.EventResponse); err != nil {
		return rpc.Response{}, err
	}
	if strings.TrimSpace(response.Command) == "" {
		return rpc.Response{}, fmt.Errorf("%w: response missing command", ErrProtocolViolation)
	}
	return response, nil
}

func validatePromptRequest(request PromptRequest, allowStreamingBehavior bool) error {
	if strings.TrimSpace(request.Message) == "" {
		return errors.New("message is required")
	}
	if !allowStreamingBehavior && request.StreamingBehavior != "" {
		return errors.New("streaming behavior is not allowed for this command")
	}
	switch request.StreamingBehavior {
	case "", StreamingBehaviorSteer, StreamingBehaviorFollowUp:
	default:
		return fmt.Errorf("invalid streaming behavior %q", request.StreamingBehavior)
	}
	return validateImages(request.Images)
}

func validateImages(images []ImageContent) error {
	for index, image := range images {
		if strings.TrimSpace(image.Data) == "" {
			return fmt.Errorf("images[%d].data is required", index)
		}
		if strings.TrimSpace(image.MIMEType) == "" {
			return fmt.Errorf("images[%d].mimeType is required", index)
		}
	}
	return nil
}

func encodeImages(images []ImageContent) []map[string]string {
	encoded := make([]map[string]string, 0, len(images))
	for _, image := range images {
		encoded = append(encoded, map[string]string{
			"type":     "image",
			"data":     image.Data,
			"mimeType": image.MIMEType,
		})
	}
	return encoded
}
