package pi

import (
	"fmt"
	"strings"
)

const (
	DefaultProvider       = "anthropic"
	DefaultModel          = "claude-opus-4-5"
	DefaultThinking       = "high"
	DefaultDumbThinking   = "low"
	DefaultFastModel      = "claude-haiku-4-5"
	DefaultCodingProvider = "openai-codex"
	DefaultCodingModel    = "gpt-5.2-codex"
	DefaultCodingThinking = "high"
)

type Mode string

const (
	ModeSmart   Mode = "smart"
	ModeDumb    Mode = "dumb"
	ModeFast    Mode = "fast"
	ModeCoding  Mode = "coding"
	ModeDragons Mode = "dragons"
)

type DragonsOptions struct {
	Provider string
	Model    string
	Thinking string
}

type SessionOptions struct {
	AppName      string
	WorkDir      string
	SystemPrompt string
	Mode         Mode
	Dragons      DragonsOptions
	SessionName  string
}

type OneShotOptions struct {
	AppName      string
	WorkDir      string
	SystemPrompt string
	Mode         Mode
	Dragons      DragonsOptions
}

func DefaultSessionOptions() SessionOptions {
	return SessionOptions{Mode: ModeSmart}
}

func DefaultOneShotOptions() OneShotOptions {
	return OneShotOptions{Mode: ModeSmart}
}

func normalizeSessionOptions(options SessionOptions) (SessionOptions, error) {
	if options.AppName == "" {
		options.AppName = "pi-golang"
	}
	if options.Mode == "" {
		options.Mode = ModeSmart
	}
	if err := validateMode(options.Mode, options.Dragons); err != nil {
		return options, err
	}
	options.Dragons = trimDragons(options.Dragons)
	options.SessionName = strings.TrimSpace(options.SessionName)
	return options, nil
}

func normalizeOneShotOptions(options OneShotOptions) (OneShotOptions, error) {
	if options.AppName == "" {
		options.AppName = "pi-golang"
	}
	if options.Mode == "" {
		options.Mode = ModeSmart
	}
	if err := validateMode(options.Mode, options.Dragons); err != nil {
		return options, err
	}
	options.Dragons = trimDragons(options.Dragons)
	return options, nil
}

func validateMode(mode Mode, dragons DragonsOptions) error {
	switch mode {
	case ModeSmart, ModeDumb, ModeFast, ModeCoding, ModeDragons:
	default:
		return fmt.Errorf("invalid mode %q", mode)
	}

	if mode != ModeDragons {
		if strings.TrimSpace(dragons.Provider) != "" ||
			strings.TrimSpace(dragons.Model) != "" ||
			strings.TrimSpace(dragons.Thinking) != "" {
			return fmt.Errorf("dragons options require mode %q", ModeDragons)
		}
		return nil
	}

	if strings.TrimSpace(dragons.Provider) == "" {
		return fmt.Errorf("dragons provider is required")
	}
	if strings.TrimSpace(dragons.Model) == "" {
		return fmt.Errorf("dragons model is required")
	}
	if strings.TrimSpace(dragons.Thinking) == "" {
		return fmt.Errorf("dragons thinking is required")
	}
	return nil
}

func trimDragons(dragons DragonsOptions) DragonsOptions {
	dragons.Provider = strings.TrimSpace(dragons.Provider)
	dragons.Model = strings.TrimSpace(dragons.Model)
	dragons.Thinking = strings.TrimSpace(dragons.Thinking)
	return dragons
}
