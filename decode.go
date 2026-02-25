package pi

import (
	"encoding/json"

	"github.com/joshp123/pi-golang/internal/sdk"
)

func DecodeAgentEnd(raw json.RawMessage) (AgentEndEvent, error) {
	return sdk.DecodeAgentEnd(raw)
}

func DecodeMessageUpdate(raw json.RawMessage) (MessageUpdateEvent, error) {
	return sdk.DecodeMessageUpdate(raw)
}

func DecodeAutoCompactionStart(raw json.RawMessage) (AutoCompactionStartEvent, error) {
	return sdk.DecodeAutoCompactionStart(raw)
}

func DecodeAutoCompactionEnd(raw json.RawMessage) (AutoCompactionEndEvent, error) {
	return sdk.DecodeAutoCompactionEnd(raw)
}

func DecodeAutoRetryStart(raw json.RawMessage) (AutoRetryStartEvent, error) {
	return sdk.DecodeAutoRetryStart(raw)
}

func DecodeAutoRetryEnd(raw json.RawMessage) (AutoRetryEndEvent, error) {
	return sdk.DecodeAutoRetryEnd(raw)
}

func DecodeTerminalOutcome(raw json.RawMessage) (TerminalOutcome, error) {
	return sdk.DecodeTerminalOutcome(raw)
}
