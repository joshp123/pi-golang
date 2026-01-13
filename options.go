package pi

const (
	DefaultProvider = "anthropic"
	DefaultModel    = "claude-opus-4-5"
	DefaultThinking = "high"
)

type Options struct {
	PiPath           string
	NodePath         string
	WorkDir          string
	AgentDir         string
	Env              map[string]string
	EnvAllowlist     []string
	EnvAllowPrefixes []string
	Args             []string
	Provider         string
	Model            string
	Thinking         string
	UsePiDefaults    bool
}

func DefaultOptions() Options {
	return Options{}
}

func (options Options) withDefaults() Options {
	if options.UsePiDefaults {
		return options
	}
	if options.Provider == "" {
		options.Provider = DefaultProvider
	}
	if options.Model == "" {
		options.Model = DefaultModel
	}
	if options.Thinking == "" {
		options.Thinking = DefaultThinking
	}
	return options
}
