package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var DefaultEnvAllowlist = []string{
	"HOME",
	"PATH",
	"USER",
	"LOGNAME",
	"LANG",
	"LC_ALL",
	"LC_CTYPE",
	"TERM",
	"SHELL",
	"TMPDIR",
	"TZ",
	"ANTHROPIC_API_KEY",
	"ANTHROPIC_OAUTH_TOKEN",
	"ANTHROPIC_TOKEN_FILE",
	"OPENAI_API_KEY",
	"PI_CODING_AGENT_DIR",
	"GEMINI_API_KEY",
	"MISTRAL_API_KEY",
	"GROQ_API_KEY",
	"CEREBRAS_API_KEY",
	"XAI_API_KEY",
	"OPENROUTER_API_KEY",
	"ZAI_API_KEY",
	"MINIMAX_API_KEY",
	"GOOGLE_CLOUD_PROJECT",
	"GOOGLE_CLOUD_PROJECT_ID",
	"AWS_PROFILE",
	"AWS_ACCESS_KEY_ID",
	"AWS_SECRET_ACCESS_KEY",
	"AWS_BEARER_TOKEN_BEDROCK",
	"AWS_REGION",
}

var DefaultEnvAllowPrefixes = []string{
	"XDG_",
	"LC_",
}

func buildEnv(appName string, inheritEnvironment bool, seedAuthFromHome bool, auth ProviderAuth, explicitValues map[string]string) ([]string, error) {
	allowed := map[string]bool{}
	for _, key := range DefaultEnvAllowlist {
		allowed[key] = true
	}

	result := map[string]string{}
	if inheritEnvironment {
		for _, entry := range os.Environ() {
			key, value, ok := strings.Cut(entry, "=")
			if !ok {
				continue
			}
			if allowed[key] || hasAllowedPrefix(key, DefaultEnvAllowPrefixes) {
				result[key] = value
			}
		}
	}

	authEnv, err := authEnvironment(auth)
	if err != nil {
		return nil, err
	}
	for key, value := range authEnv {
		result[key] = value
	}

	for key, value := range explicitValues {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if isCredentialEnvironmentKey(key) {
			return nil, fmt.Errorf("credential %s must be provided via options.Auth", key)
		}
		result[key] = value
	}

	if strings.TrimSpace(result["PI_CODING_AGENT_DIR"]) == "" {
		agentDir, err := resolveAgentDir(appName, seedAuthFromHome)
		if err != nil {
			return nil, err
		}
		result["PI_CODING_AGENT_DIR"] = agentDir
	}

	return mapToEnvSlice(result), nil
}

func authEnvironment(auth ProviderAuth) (map[string]string, error) {
	values := map[string]string{}

	if err := setCredentialEnvironment(values, "ANTHROPIC_API_KEY", auth.Anthropic.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "ANTHROPIC_OAUTH_TOKEN", auth.Anthropic.OAuthToken); err != nil {
		return nil, err
	}
	if strings.TrimSpace(auth.Anthropic.TokenFilePath) != "" {
		values["ANTHROPIC_TOKEN_FILE"] = strings.TrimSpace(auth.Anthropic.TokenFilePath)
	}

	if err := setCredentialEnvironment(values, "OPENAI_API_KEY", auth.OpenAI.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "GEMINI_API_KEY", auth.Gemini.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "MISTRAL_API_KEY", auth.Mistral.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "GROQ_API_KEY", auth.Groq.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "CEREBRAS_API_KEY", auth.Cerebras.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "XAI_API_KEY", auth.XAI.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "OPENROUTER_API_KEY", auth.OpenRouter.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "ZAI_API_KEY", auth.ZAI.APIKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "MINIMAX_API_KEY", auth.Minimax.APIKey); err != nil {
		return nil, err
	}

	if err := setCredentialEnvironment(values, "AWS_PROFILE", auth.Bedrock.Profile); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "AWS_ACCESS_KEY_ID", auth.Bedrock.AccessKeyID); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "AWS_SECRET_ACCESS_KEY", auth.Bedrock.SecretAccessKey); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "AWS_BEARER_TOKEN_BEDROCK", auth.Bedrock.BearerToken); err != nil {
		return nil, err
	}
	if err := setCredentialEnvironment(values, "AWS_REGION", auth.Bedrock.Region); err != nil {
		return nil, err
	}

	return values, nil
}

func setCredentialEnvironment(values map[string]string, key string, credential Credential) error {
	resolved, ok, err := resolveCredentialValue(credential)
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}
	if ok {
		values[key] = resolved
	}
	return nil
}

func resolveCredentialValue(credential Credential) (string, bool, error) {
	hasValue := strings.TrimSpace(credential.Value) != ""
	hasFile := strings.TrimSpace(credential.File) != ""
	if hasValue && hasFile {
		return "", false, fmt.Errorf("set exactly one of Value or File")
	}
	if hasValue {
		return strings.TrimSpace(credential.Value), true, nil
	}
	if !hasFile {
		return "", false, nil
	}
	filePath := strings.TrimSpace(credential.File)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", false, err
	}
	resolved := strings.TrimSpace(string(data))
	if resolved == "" {
		return "", false, fmt.Errorf("file is empty")
	}
	return resolved, true, nil
}

func isCredentialEnvironmentKey(key string) bool {
	switch strings.TrimSpace(key) {
	case "ANTHROPIC_API_KEY",
		"ANTHROPIC_OAUTH_TOKEN",
		"ANTHROPIC_TOKEN_FILE",
		"OPENAI_API_KEY",
		"GEMINI_API_KEY",
		"MISTRAL_API_KEY",
		"GROQ_API_KEY",
		"CEREBRAS_API_KEY",
		"XAI_API_KEY",
		"OPENROUTER_API_KEY",
		"ZAI_API_KEY",
		"MINIMAX_API_KEY",
		"AWS_PROFILE",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_BEARER_TOKEN_BEDROCK",
		"AWS_REGION":
		return true
	default:
		return false
	}
}

func hasAllowedPrefix(key string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func resolveAgentDir(appName string, seedAuth bool) (string, error) {
	if strings.TrimSpace(appName) == "" {
		appName = "pi-golang"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	name := strings.TrimPrefix(appName, ".")
	root := filepath.Join(home, "."+name)
	agentDir := filepath.Join(root, "pi-agent")
	if err := os.MkdirAll(agentDir, 0o700); err != nil {
		return "", err
	}

	if seedAuth {
		if err := seedAuthFiles(agentDir); err != nil {
			return "", err
		}
	}

	return agentDir, nil
}

func seedAuthFiles(agentDir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	primaryDir := filepath.Join(home, ".pi", "agent")
	authSource := filepath.Join(primaryDir, "auth.json")
	authDest := filepath.Join(agentDir, "auth.json")
	if fileExists(authSource) {
		if err := copyIfNewer(authSource, authDest, 0o600); err != nil {
			return err
		}
	}

	oauthSource := filepath.Join(primaryDir, "oauth.json")
	oauthDest := filepath.Join(agentDir, "oauth.json")
	if fileExists(oauthSource) {
		if err := copyIfNewer(oauthSource, oauthDest, 0o600); err != nil {
			return err
		}
	}

	return nil
}

func copyIfNewer(source string, dest string, mode os.FileMode) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	if destInfo, err := os.Stat(dest); err == nil {
		if !sourceInfo.ModTime().After(destInfo.ModTime()) {
			return nil
		}
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dest, data, mode); err != nil {
		return err
	}
	return nil
}

func mapToEnvSlice(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key+"="+values[key])
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
