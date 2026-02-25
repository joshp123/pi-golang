package pi

import (
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

func buildEnv(appName string) ([]string, error) {
	allowed := map[string]bool{}
	for _, key := range DefaultEnvAllowlist {
		allowed[key] = true
	}

	result := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if allowed[key] || hasAllowedPrefix(key, DefaultEnvAllowPrefixes) {
			result[key] = value
		}
	}

	// Only set PI_CODING_AGENT_DIR if not already in environment
	if result["PI_CODING_AGENT_DIR"] == "" {
		agentDir, err := resolveAgentDir(appName)
		if err != nil {
			return nil, err
		}
		result["PI_CODING_AGENT_DIR"] = agentDir
	}

	return mapToEnvSlice(result), nil
}

func hasAllowedPrefix(key string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func resolveAgentDir(appName string) (string, error) {
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

	if err := seedAuthFiles(agentDir); err != nil {
		return "", err
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
		return nil
	}

	for _, legacyPath := range legacyOAuthPaths(home) {
		if fileExists(legacyPath) {
			return copyIfNewer(legacyPath, oauthDest, 0o600)
		}
	}

	return nil
}

func legacyOAuthPaths(home string) []string {
	paths := []string{}
	if override := os.Getenv("PI_CODING_AGENT_DIR"); override != "" {
		paths = append(paths, filepath.Join(expandHome(override, home), "oauth.json"))
	}
	paths = append(paths,
		filepath.Join(home, ".clawdis", "credentials", "oauth.json"),
		filepath.Join(home, ".claude", "oauth.json"),
		filepath.Join(home, ".config", "claude", "oauth.json"),
		filepath.Join(home, ".config", "anthropic", "oauth.json"),
	)
	return paths
}

func expandHome(path string, home string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	return filepath.Join(home, path[1:])
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
