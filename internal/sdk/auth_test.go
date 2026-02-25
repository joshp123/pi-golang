package sdk

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateProviderAuthAnthropicMissing(t *testing.T) {
	err := validateProviderAuth("anthropic", ProviderAuth{})
	if err == nil {
		t.Fatal("expected missing auth error")
	}
	var missing *MissingProviderAuthError
	if !errors.As(err, &missing) {
		t.Fatalf("expected MissingProviderAuthError, got %T: %v", err, err)
	}
}

func TestValidateProviderAuthAnthropicPresentValue(t *testing.T) {
	auth := ProviderAuth{}
	auth.Anthropic.APIKey = Credential{Value: "abc"}
	if err := validateProviderAuth("anthropic", auth); err != nil {
		t.Fatalf("validateProviderAuth failed: %v", err)
	}
}

func TestValidateProviderAuthOpenAIPresentFile(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "openai.txt")
	writeSecretFile(t, keyFile, "token")

	auth := ProviderAuth{}
	auth.OpenAI.APIKey = Credential{File: keyFile}
	if err := validateProviderAuth("openai-codex", auth); err != nil {
		t.Fatalf("validateProviderAuth failed: %v", err)
	}
}

func TestValidateProviderAuthRejectsMissingFile(t *testing.T) {
	auth := ProviderAuth{}
	auth.OpenAI.APIKey = Credential{File: "/tmp/does-not-exist-pi-golang"}
	if err := validateProviderAuth("openai-codex", auth); err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestValidateProviderAuthRejectsValueAndFile(t *testing.T) {
	auth := ProviderAuth{}
	auth.OpenAI.APIKey = Credential{Value: "token", File: "/tmp/anything"}
	if err := validateProviderAuth("openai-codex", auth); err == nil {
		t.Fatal("expected value+file error")
	}
}

func TestValidateProviderAuthUnknownRequiresAnyCredential(t *testing.T) {
	if err := validateProviderAuth("unknown-provider", ProviderAuth{}); err == nil {
		t.Fatal("expected missing auth error for unknown provider")
	}

	auth := ProviderAuth{}
	auth.Groq.APIKey = Credential{Value: "k"}
	if err := validateProviderAuth("unknown-provider", auth); err != nil {
		t.Fatalf("expected unknown provider with any credential to pass, got %v", err)
	}
}

func writeSecretFile(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value+"\n"), 0o600); err != nil {
		t.Fatalf("write secret file: %v", err)
	}
}
