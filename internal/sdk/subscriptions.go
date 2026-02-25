package sdk

import (
	"encoding/json"
	"fmt"

	"github.com/joshp123/pi-golang/internal/stream"
)

func newEventHub() *stream.Hub[Event] {
	return stream.NewHub[Event](
		ErrClientClosed,
		func(event Event) string { return event.Type },
		EventTypeSubscriptionDrop,
		func(mode stream.Mode, droppedType string) Event {
			return newSubscriptionDropEvent(fromStreamMode(mode), droppedType)
		},
	)
}

func validateSubscriptionPolicy(policy SubscriptionPolicy) error {
	if policy.Buffer <= 0 {
		return fmt.Errorf("%w: buffer must be > 0", ErrInvalidSubscriptionPolicy)
	}
	switch policy.Mode {
	case SubscriptionModeDrop, SubscriptionModeBlock, SubscriptionModeRing:
		return nil
	default:
		return fmt.Errorf("%w: unsupported mode %q", ErrInvalidSubscriptionPolicy, policy.Mode)
	}
}

func toStreamPolicy(policy SubscriptionPolicy) stream.Policy {
	return stream.Policy{
		Buffer:        policy.Buffer,
		Mode:          toStreamMode(policy.Mode),
		EmitDropEvent: policy.EmitDropEvent,
	}
}

func toStreamMode(mode SubscriptionMode) stream.Mode {
	switch mode {
	case SubscriptionModeBlock:
		return stream.ModeBlock
	case SubscriptionModeRing:
		return stream.ModeRing
	default:
		return stream.ModeDrop
	}
}

func fromStreamMode(mode stream.Mode) SubscriptionMode {
	switch mode {
	case stream.ModeBlock:
		return SubscriptionModeBlock
	case stream.ModeRing:
		return SubscriptionModeRing
	default:
		return SubscriptionModeDrop
	}
}

func newSubscriptionDropEvent(mode SubscriptionMode, droppedType string) Event {
	raw, _ := json.Marshal(map[string]any{
		"type":        EventTypeSubscriptionDrop,
		"mode":        mode,
		"droppedType": droppedType,
	})
	return Event{Type: EventTypeSubscriptionDrop, Raw: raw}
}
