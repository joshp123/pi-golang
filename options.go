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

type Options struct {
	AppName      string
	WorkDir      string
	SystemPrompt string
	Mode         Mode
	Dragons      DragonsOptions
}

func DefaultOptions() Options {
	return Options{Mode: ModeSmart}
}

func (options Options) withDefaults() Options {
	if options.AppName == "" {
		options.AppName = "pi-golang"
	}
	if options.Mode == "" {
		options.Mode = ModeSmart
	}
	return options
}

func (options Options) validate() error {
	switch options.Mode {
	case ModeSmart, ModeDumb, ModeFast, ModeCoding, ModeDragons:
	default:
		return fmt.Errorf("invalid mode %q", options.Mode)
	}

	if options.Mode != ModeDragons {
		if strings.TrimSpace(options.Dragons.Provider) != "" ||
			strings.TrimSpace(options.Dragons.Model) != "" ||
			strings.TrimSpace(options.Dragons.Thinking) != "" {
			return fmt.Errorf("dragons options require mode %q", ModeDragons)
		}
		return nil
	}

	if strings.TrimSpace(options.Dragons.Provider) == "" {
		return fmt.Errorf("dragons provider is required")
	}
	if strings.TrimSpace(options.Dragons.Model) == "" {
		return fmt.Errorf("dragons model is required")
	}
	if strings.TrimSpace(options.Dragons.Thinking) == "" {
		return fmt.Errorf("dragons thinking is required")
	}
	return nil
}
