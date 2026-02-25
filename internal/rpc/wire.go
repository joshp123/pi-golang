package rpc

import "encoding/json"

// Command is a raw upstream RPC command payload.
type Command map[string]any

// Response is a raw upstream RPC response frame.
type Response struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Command string          `json:"command,omitempty"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

const (
	CommandPrompt     = "prompt"
	CommandSteer      = "steer"
	CommandFollowUp   = "follow_up"
	CommandAbort      = "abort"
	CommandGetState   = "get_state"
	CommandNewSession = "new_session"
	CommandCompact    = "compact"
	CommandExportHTML = "export_html"
)

const (
	EventResponse           = "response"
	EventParseError         = "parse_error"
	EventResponseParseError = "response_parse_error"
)
