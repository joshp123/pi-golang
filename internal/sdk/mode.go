package sdk

type modelConfig struct {
	provider string
	model    string
	thinking string
}

func resolveModelConfig(mode Mode, dragons DragonsOptions) (modelConfig, error) {
	if err := validateMode(mode, dragons); err != nil {
		return modelConfig{}, err
	}
	dragons = trimDragons(dragons)

	switch mode {
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
			provider: dragons.Provider,
			model:    dragons.Model,
			thinking: dragons.Thinking,
		}, nil
	default:
		return modelConfig{}, validateMode(mode, dragons)
	}
}
