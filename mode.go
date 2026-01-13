package pi

import "strings"

type modelConfig struct {
	provider string
	model    string
	thinking string
}

func resolveModelConfig(options Options) (modelConfig, error) {
	options = options.withDefaults()
	if err := options.validate(); err != nil {
		return modelConfig{}, err
	}

	switch options.Mode {
	case ModeSmart:
		return modelConfig{
			provider: DefaultProvider,
			model:    DefaultModel,
			thinking: DefaultThinking,
		}, nil
	case ModeDumb:
		return modelConfig{
			provider: DefaultProvider,
			model:    DefaultModel,
			thinking: DefaultDumbThinking,
		}, nil
	case ModeFast:
		return modelConfig{
			provider: DefaultProvider,
			model:    DefaultFastModel,
			thinking: DefaultDumbThinking,
		}, nil
	case ModeCoding:
		return modelConfig{
			provider: DefaultCodingProvider,
			model:    DefaultCodingModel,
			thinking: DefaultCodingThinking,
		}, nil
	case ModeDragons:
		return modelConfig{
			provider: strings.TrimSpace(options.Dragons.Provider),
			model:    strings.TrimSpace(options.Dragons.Model),
			thinking: strings.TrimSpace(options.Dragons.Thinking),
		}, nil
	default:
		return modelConfig{}, options.validate()
	}
}
