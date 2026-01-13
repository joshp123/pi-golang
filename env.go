package pi

import (
	"os"
	"path/filepath"
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
}

var DefaultEnvAllowPrefixes = []string{
	"XDG_",
	"LC_",
}

func buildEnv(options Options) ([]string, error) {
	allowlist := options.EnvAllowlist
	if allowlist == nil {
		allowlist = DefaultEnvAllowlist
	}
	allowPrefixes := options.EnvAllowPrefixes
	if allowPrefixes == nil {
		allowPrefixes = DefaultEnvAllowPrefixes
	}

	allowed := map[string]bool{}
	for _, key := range allowlist {
		allowed[key] = true
	}

	result := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if allowed[key] || hasAllowedPrefix(key, allowPrefixes) {
			result[key] = value
		}
	}

	agentDir, err := resolveAgentDir(options)
	if err != nil {
		return nil, err
	}
	result["PI_CODING_AGENT_DIR"] = agentDir

	for key, value := range options.Env {
		result[key] = value
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

func resolveAgentDir(options Options) (string, error) {
	if options.AgentDir != "" {
		if err := os.MkdirAll(options.AgentDir, 0o700); err != nil {
			return "", err
		}
		return options.AgentDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	base := filepath.Join(home, ".pi-golang", "agent")
	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", err
	}

	if err := seedAuthFiles(base); err != nil {
		return "", err
	}

	return base, nil
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
	out := make([]string, 0, len(values))
	for key, value := range values {
		out = append(out, key+"="+value)
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
