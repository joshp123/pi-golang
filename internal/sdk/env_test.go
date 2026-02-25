package sdk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildEnvRespectsExplicitPICodingAgentDir(t *testing.T) {
	customPath := "/custom/agent/dir"
	env, err := buildEnv("test-app", false, false, ProviderAuth{}, map[string]string{"PI_CODING_AGENT_DIR": customPath})
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}

	values := envToMap(env)
	if values["PI_CODING_AGENT_DIR"] != customPath {
		t.Errorf("PI_CODING_AGENT_DIR not respected: got %q, want %q", values["PI_CODING_AGENT_DIR"], customPath)
	}
}

func TestBuildEnvSetsDefaultPICodingAgentDir(t *testing.T) {
	env, err := buildEnv("test-app", false, false, ProviderAuth{}, map[string]string{})
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}

	values := envToMap(env)
	found := values["PI_CODING_AGENT_DIR"]
	if found == "" {
		t.Fatal("PI_CODING_AGENT_DIR not set")
	}
}

func TestBuildEnvInheritAllowlistedVars(t *testing.T) {
	env, err := buildEnv("test-app", true, false, ProviderAuth{}, map[string]string{})
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}

	values := envToMap(env)
	if values["HOME"] == "" {
		t.Error("HOME not in inherited env")
	}
	if values["PATH"] == "" {
		t.Error("PATH not in inherited env")
	}
}

func TestBuildEnvDoesNotInheritHostWhenDisabled(t *testing.T) {
	env, err := buildEnv("test-app", false, false, ProviderAuth{}, map[string]string{"FOO": "bar"})
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}

	values := envToMap(env)
	if values["FOO"] != "bar" {
		t.Fatalf("explicit env var missing")
	}
	if _, exists := values["HOME"]; exists {
		t.Fatal("HOME should not be inherited when inherit=false")
	}
	if _, exists := values["PATH"]; exists {
		t.Fatal("PATH should not be inherited when inherit=false")
	}
}

func TestBuildEnvMapsAuthCredentialValue(t *testing.T) {
	auth := ProviderAuth{}
	auth.OpenAI.APIKey = Credential{Value: "openai-token"}
	env, err := buildEnv("test-app", false, false, auth, map[string]string{})
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}
	values := envToMap(env)
	if values["OPENAI_API_KEY"] != "openai-token" {
		t.Fatalf("expected OPENAI_API_KEY from auth, got %q", values["OPENAI_API_KEY"])
	}
}

func TestBuildEnvMapsAuthCredentialFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "anthropic.txt")
	if err := os.WriteFile(path, []byte("secret-token\n"), 0o600); err != nil {
		t.Fatalf("write secret file: %v", err)
	}
	auth := ProviderAuth{}
	auth.Anthropic.APIKey = Credential{File: path}
	env, err := buildEnv("test-app", false, false, auth, map[string]string{})
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}
	values := envToMap(env)
	if values["ANTHROPIC_API_KEY"] != "secret-token" {
		t.Fatalf("expected env file value, got %q", values["ANTHROPIC_API_KEY"])
	}
}

func TestBuildEnvRejectsCredentialOverrideInEnvironmentMap(t *testing.T) {
	auth := ProviderAuth{}
	auth.OpenAI.APIKey = Credential{Value: "a"}
	_, err := buildEnv("test-app", false, false, auth, map[string]string{"OPENAI_API_KEY": "b"})
	if err == nil {
		t.Fatal("expected credential override error")
	}
}

func envToMap(env []string) map[string]string {
	values := map[string]string{}
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		values[key] = value
	}
	return values
}
