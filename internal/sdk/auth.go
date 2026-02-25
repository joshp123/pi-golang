package sdk

import (
	"fmt"
	"os"
	"strings"
)

// MissingProviderAuthError indicates no usable credential was configured for the selected provider.
type MissingProviderAuthError struct {
	Provider string
	Required string
}

func (err *MissingProviderAuthError) Error() string {
	if err == nil {
		return ""
	}
	provider := strings.TrimSpace(err.Provider)
	if provider == "" {
		provider = "unknown"
	}
	required := strings.TrimSpace(err.Required)
	if required == "" {
		required = "provider credential"
	}
	return fmt.Sprintf("missing auth for provider %q: set %s", provider, required)
}

func validateProviderAuth(provider string, auth ProviderAuth) error {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	switch {
	case normalized == "":
		return nil
	case strings.Contains(normalized, "anthropic"):
		apiKeyPresent, err := credentialPresent(auth.Anthropic.APIKey)
		if err != nil {
			return fmt.Errorf("Anthropic.APIKey: %w", err)
		}
		oauthTokenPresent, err := credentialPresent(auth.Anthropic.OAuthToken)
		if err != nil {
			return fmt.Errorf("Anthropic.OAuthToken: %w", err)
		}
		tokenFilePresent, err := tokenFilePresent(auth.Anthropic.TokenFilePath)
		if err != nil {
			return fmt.Errorf("Anthropic.TokenFilePath: %w", err)
		}
		if apiKeyPresent || oauthTokenPresent || tokenFilePresent {
			return nil
		}
		return &MissingProviderAuthError{Provider: provider, Required: "Anthropic.APIKey or Anthropic.OAuthToken or Anthropic.TokenFilePath"}
	case strings.Contains(normalized, "openai"):
		return requireCredential(provider, "OpenAI.APIKey", auth.OpenAI.APIKey)
	case strings.Contains(normalized, "gemini") || strings.Contains(normalized, "google"):
		return requireCredential(provider, "Gemini.APIKey", auth.Gemini.APIKey)
	case strings.Contains(normalized, "mistral"):
		return requireCredential(provider, "Mistral.APIKey", auth.Mistral.APIKey)
	case strings.Contains(normalized, "groq"):
		return requireCredential(provider, "Groq.APIKey", auth.Groq.APIKey)
	case strings.Contains(normalized, "cerebras"):
		return requireCredential(provider, "Cerebras.APIKey", auth.Cerebras.APIKey)
	case strings.Contains(normalized, "openrouter"):
		return requireCredential(provider, "OpenRouter.APIKey", auth.OpenRouter.APIKey)
	case strings.Contains(normalized, "minimax"):
		return requireCredential(provider, "Minimax.APIKey", auth.Minimax.APIKey)
	case normalized == "xai" || strings.Contains(normalized, "x-ai") || strings.Contains(normalized, "grok"):
		return requireCredential(provider, "XAI.APIKey", auth.XAI.APIKey)
	case normalized == "zai" || strings.Contains(normalized, "glm"):
		return requireCredential(provider, "ZAI.APIKey", auth.ZAI.APIKey)
	case strings.Contains(normalized, "bedrock") || strings.Contains(normalized, "aws"):
		return requireBedrockCredential(provider, auth.Bedrock)
	default:
		present, err := anyProviderCredentialPresent(auth)
		if err != nil {
			return err
		}
		if present {
			return nil
		}
		return &MissingProviderAuthError{Provider: provider, Required: "at least one credential in options.Auth"}
	}
}

func requireCredential(provider string, label string, credential Credential) error {
	present, err := credentialPresent(credential)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if present {
		return nil
	}
	return &MissingProviderAuthError{Provider: provider, Required: label}
}

func requireBedrockCredential(provider string, auth BedrockAuth) error {
	profilePresent, err := credentialPresent(auth.Profile)
	if err != nil {
		return fmt.Errorf("Bedrock.Profile: %w", err)
	}
	accessKeyPresent, err := credentialPresent(auth.AccessKeyID)
	if err != nil {
		return fmt.Errorf("Bedrock.AccessKeyID: %w", err)
	}
	secretKeyPresent, err := credentialPresent(auth.SecretAccessKey)
	if err != nil {
		return fmt.Errorf("Bedrock.SecretAccessKey: %w", err)
	}
	bearerPresent, err := credentialPresent(auth.BearerToken)
	if err != nil {
		return fmt.Errorf("Bedrock.BearerToken: %w", err)
	}
	if profilePresent || bearerPresent || (accessKeyPresent && secretKeyPresent) {
		return nil
	}
	return &MissingProviderAuthError{Provider: provider, Required: "Bedrock.Profile or Bedrock.BearerToken or Bedrock.AccessKeyID+Bedrock.SecretAccessKey"}
}

func anyProviderCredentialPresent(auth ProviderAuth) (bool, error) {
	checks := []struct {
		label      string
		credential Credential
	}{
		{"Anthropic.APIKey", auth.Anthropic.APIKey},
		{"Anthropic.OAuthToken", auth.Anthropic.OAuthToken},
		{"OpenAI.APIKey", auth.OpenAI.APIKey},
		{"Gemini.APIKey", auth.Gemini.APIKey},
		{"Mistral.APIKey", auth.Mistral.APIKey},
		{"Groq.APIKey", auth.Groq.APIKey},
		{"Cerebras.APIKey", auth.Cerebras.APIKey},
		{"XAI.APIKey", auth.XAI.APIKey},
		{"OpenRouter.APIKey", auth.OpenRouter.APIKey},
		{"ZAI.APIKey", auth.ZAI.APIKey},
		{"Minimax.APIKey", auth.Minimax.APIKey},
		{"Bedrock.Profile", auth.Bedrock.Profile},
		{"Bedrock.AccessKeyID", auth.Bedrock.AccessKeyID},
		{"Bedrock.SecretAccessKey", auth.Bedrock.SecretAccessKey},
		{"Bedrock.BearerToken", auth.Bedrock.BearerToken},
		{"Bedrock.Region", auth.Bedrock.Region},
	}
	for _, check := range checks {
		present, err := credentialPresent(check.credential)
		if err != nil {
			return false, fmt.Errorf("%s: %w", check.label, err)
		}
		if present {
			return true, nil
		}
	}
	tokenFilePresent, err := tokenFilePresent(auth.Anthropic.TokenFilePath)
	if err != nil {
		return false, fmt.Errorf("Anthropic.TokenFilePath: %w", err)
	}
	return tokenFilePresent, nil
}

func credentialPresent(credential Credential) (bool, error) {
	hasValue := strings.TrimSpace(credential.Value) != ""
	hasFile := strings.TrimSpace(credential.File) != ""
	if hasValue && hasFile {
		return false, fmt.Errorf("set exactly one of Value or File")
	}
	if hasValue {
		return true, nil
	}
	if !hasFile {
		return false, nil
	}
	filePath := strings.TrimSpace(credential.File)
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("file path points to directory")
	}
	if info.Size() == 0 {
		return false, fmt.Errorf("file is empty")
	}
	return true, nil
}

func tokenFilePresent(path string) (bool, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return false, nil
	}
	info, err := os.Stat(trimmed)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("path points to directory")
	}
	if info.Size() == 0 {
		return false, fmt.Errorf("file is empty")
	}
	return true, nil
}
