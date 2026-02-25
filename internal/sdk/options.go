package sdk

import (
	"fmt"
	"strings"
)

const (
	DefaultProvider       = "anthropic"
	DefaultModel          = "claude-opus-4-6"
	DefaultThinking       = "high"
	DefaultDumbThinking   = "low"
	DefaultFastModel      = "claude-haiku-4-5"
	DefaultCodingProvider = "openai-codex"
	DefaultCodingModel    = "gpt-5.3-codex"
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

type Credential struct {
	Value string
	File  string
}

type APIKeyAuth struct {
	APIKey Credential
}

type AnthropicAuth struct {
	APIKey        Credential
	OAuthToken    Credential
	TokenFilePath string
}

type BedrockAuth struct {
	Profile         Credential
	AccessKeyID     Credential
	SecretAccessKey Credential
	BearerToken     Credential
	Region          Credential
}

type ProviderAuth struct {
	Anthropic  AnthropicAuth
	OpenAI     APIKeyAuth
	Gemini     APIKeyAuth
	Mistral    APIKeyAuth
	Groq       APIKeyAuth
	Cerebras   APIKeyAuth
	XAI        APIKeyAuth
	OpenRouter APIKeyAuth
	ZAI        APIKeyAuth
	Minimax    APIKeyAuth
	Bedrock    BedrockAuth
}

type SessionOptions struct {
	AppName            string
	WorkDir            string
	SystemPrompt       string
	Mode               Mode
	Dragons            DragonsOptions
	SessionName        string
	Auth               ProviderAuth
	Environment        map[string]string
	InheritEnvironment bool
	SeedAuthFromHome   bool
	CompactionPrompt   string
}

type OneShotOptions struct {
	AppName            string
	WorkDir            string
	SystemPrompt       string
	Mode               Mode
	Dragons            DragonsOptions
	Auth               ProviderAuth
	Environment        map[string]string
	InheritEnvironment bool
	SeedAuthFromHome   bool
	CompactionPrompt   string
}

func DefaultSessionOptions() SessionOptions {
	return SessionOptions{Mode: ModeSmart, InheritEnvironment: false, SeedAuthFromHome: true}
}

func DefaultOneShotOptions() OneShotOptions {
	return OneShotOptions{Mode: ModeSmart, InheritEnvironment: false, SeedAuthFromHome: true}
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
	options.Auth = trimProviderAuth(options.Auth)
	options.Environment = cloneStringMap(options.Environment)
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
	options.Auth = trimProviderAuth(options.Auth)
	options.Environment = cloneStringMap(options.Environment)
	return options, nil
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return map[string]string{}
	}
	copy := make(map[string]string, len(values))
	for key, value := range values {
		copy[key] = value
	}
	return copy
}

func trimProviderAuth(auth ProviderAuth) ProviderAuth {
	auth.Anthropic.APIKey = trimCredential(auth.Anthropic.APIKey)
	auth.Anthropic.OAuthToken = trimCredential(auth.Anthropic.OAuthToken)
	auth.Anthropic.TokenFilePath = strings.TrimSpace(auth.Anthropic.TokenFilePath)
	auth.OpenAI.APIKey = trimCredential(auth.OpenAI.APIKey)
	auth.Gemini.APIKey = trimCredential(auth.Gemini.APIKey)
	auth.Mistral.APIKey = trimCredential(auth.Mistral.APIKey)
	auth.Groq.APIKey = trimCredential(auth.Groq.APIKey)
	auth.Cerebras.APIKey = trimCredential(auth.Cerebras.APIKey)
	auth.XAI.APIKey = trimCredential(auth.XAI.APIKey)
	auth.OpenRouter.APIKey = trimCredential(auth.OpenRouter.APIKey)
	auth.ZAI.APIKey = trimCredential(auth.ZAI.APIKey)
	auth.Minimax.APIKey = trimCredential(auth.Minimax.APIKey)
	auth.Bedrock.Profile = trimCredential(auth.Bedrock.Profile)
	auth.Bedrock.AccessKeyID = trimCredential(auth.Bedrock.AccessKeyID)
	auth.Bedrock.SecretAccessKey = trimCredential(auth.Bedrock.SecretAccessKey)
	auth.Bedrock.BearerToken = trimCredential(auth.Bedrock.BearerToken)
	auth.Bedrock.Region = trimCredential(auth.Bedrock.Region)
	return auth
}

func trimCredential(credential Credential) Credential {
	credential.Value = strings.TrimSpace(credential.Value)
	credential.File = strings.TrimSpace(credential.File)
	return credential
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
