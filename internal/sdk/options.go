package sdk

import (
	"fmt"
	"os"
	"path/filepath"
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

type SkillsMode string

const (
	SkillsModeDisabled SkillsMode = "disabled"
	SkillsModeExplicit SkillsMode = "explicit"
	SkillsModeAmbient  SkillsMode = "ambient"
)

type SkillsOptions struct {
	Mode  SkillsMode
	Paths []string
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
	Skills             SkillsOptions
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
	Skills             SkillsOptions
	CompactionPrompt   string
}

func DefaultSessionOptions() SessionOptions {
	return SessionOptions{
		Mode:               ModeSmart,
		InheritEnvironment: false,
		SeedAuthFromHome:   true,
		Skills:             SkillsOptions{Mode: SkillsModeDisabled},
	}
}

func DefaultOneShotOptions() OneShotOptions {
	return OneShotOptions{
		Mode:               ModeSmart,
		InheritEnvironment: false,
		SeedAuthFromHome:   true,
		Skills:             SkillsOptions{Mode: SkillsModeDisabled},
	}
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
	options.WorkDir = strings.TrimSpace(options.WorkDir)
	options.SessionName = strings.TrimSpace(options.SessionName)
	options.Auth = trimProviderAuth(options.Auth)
	options.Environment = cloneStringMap(options.Environment)
	normalizedSkills, err := normalizeSkillsOptions(options.Skills, options.WorkDir)
	if err != nil {
		return options, err
	}
	options.Skills = normalizedSkills
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
	options.WorkDir = strings.TrimSpace(options.WorkDir)
	options.Auth = trimProviderAuth(options.Auth)
	options.Environment = cloneStringMap(options.Environment)
	normalizedSkills, err := normalizeSkillsOptions(options.Skills, options.WorkDir)
	if err != nil {
		return options, err
	}
	options.Skills = normalizedSkills
	return options, nil
}

func normalizeSkillsOptions(options SkillsOptions, workDir string) (SkillsOptions, error) {
	options.Mode = SkillsMode(strings.TrimSpace(string(options.Mode)))
	if options.Mode == "" {
		options.Mode = SkillsModeDisabled
	}

	paths := make([]string, 0, len(options.Paths))
	for _, path := range options.Paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		paths = append(paths, trimmed)
	}

	switch options.Mode {
	case SkillsModeDisabled:
		if len(paths) > 0 {
			return options, fmt.Errorf("skills paths require mode %q", SkillsModeExplicit)
		}
		options.Paths = nil
		return options, nil
	case SkillsModeAmbient:
		if len(paths) > 0 {
			return options, fmt.Errorf("skills paths are not allowed in mode %q", SkillsModeAmbient)
		}
		options.Paths = nil
		return options, nil
	case SkillsModeExplicit:
		if len(paths) == 0 {
			return options, fmt.Errorf("at least one skill path is required in mode %q", SkillsModeExplicit)
		}
		baseDir, err := skillsBaseDir(workDir)
		if err != nil {
			return options, err
		}
		resolvedPaths := make([]string, 0, len(paths))
		seen := make(map[string]struct{}, len(paths))
		for _, path := range paths {
			resolvedPath := resolveSkillPath(path, baseDir)
			if _, exists := seen[resolvedPath]; exists {
				continue
			}
			info, err := os.Stat(resolvedPath)
			if err != nil {
				return options, fmt.Errorf("skill path %q: %w", path, err)
			}
			if !info.IsDir() && !strings.HasSuffix(resolvedPath, ".md") {
				return options, fmt.Errorf("skill path %q must be a directory or .md file", path)
			}
			resolvedPaths = append(resolvedPaths, resolvedPath)
			seen[resolvedPath] = struct{}{}
		}
		if len(resolvedPaths) == 0 {
			return options, fmt.Errorf("at least one skill path is required in mode %q", SkillsModeExplicit)
		}
		options.Paths = resolvedPaths
		return options, nil
	default:
		return options, fmt.Errorf("invalid skills mode %q", options.Mode)
	}
}

func skillsBaseDir(workDir string) (string, error) {
	trimmedWorkDir := strings.TrimSpace(workDir)
	if trimmedWorkDir == "" {
		return os.Getwd()
	}
	if filepath.IsAbs(trimmedWorkDir) {
		return filepath.Clean(trimmedWorkDir), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(cwd, trimmedWorkDir)), nil
}

func resolveSkillPath(path string, baseDir string) string {
	expanded := path
	home, err := os.UserHomeDir()
	if err == nil {
		switch {
		case expanded == "~":
			expanded = home
		case strings.HasPrefix(expanded, "~/"):
			expanded = filepath.Join(home, expanded[2:])
		case strings.HasPrefix(expanded, "~"):
			expanded = filepath.Join(home, expanded[1:])
		}
	}
	if filepath.IsAbs(expanded) {
		return filepath.Clean(expanded)
	}
	return filepath.Clean(filepath.Join(baseDir, expanded))
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
