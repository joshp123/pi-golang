//go:build integration

package sdk_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/joshp123/pi-golang/internal/sdk"
)

type realPIConfig struct {
	auth             sdk.ProviderAuth
	mode             sdk.Mode
	compactionPrompt string
}

func TestRealPIRunSmoke(t *testing.T) {
	config := requireRealPITestPrereqs(t)

	client, err := startRealPIClient(t, config)
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := client.Run(ctx, sdk.PromptRequest{Message: "Reply with exactly: REAL_PI_OK"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(strings.ToUpper(result.Text), "REAL_PI_OK") {
		t.Fatalf("unexpected result text: %q", result.Text)
	}
}

func TestRealPISubscribeReceivesAgentEnd(t *testing.T) {
	config := requireRealPITestPrereqs(t)

	client, err := startRealPIClient(t, config)
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	events, cancelEvents, err := client.Subscribe(sdk.SubscriptionPolicy{Buffer: 256, Mode: sdk.SubscriptionModeRing})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer cancelEvents()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := client.Prompt(ctx, sdk.PromptRequest{Message: "Reply with exactly: STREAM_OK"}); err != nil {
		t.Fatalf("Prompt failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for agent_end: %v", ctx.Err())
		case event, ok := <-events:
			if !ok {
				t.Fatal("event channel closed before agent_end")
			}
			if event.Type != sdk.EventTypeAgentEnd {
				continue
			}
			outcome, err := sdk.DecodeTerminalOutcome(event.Raw)
			if err != nil {
				t.Fatalf("DecodeTerminalOutcome failed: %v", err)
			}
			if !strings.Contains(strings.ToUpper(outcome.Text), "STREAM_OK") {
				t.Fatalf("unexpected outcome text: %q", outcome.Text)
			}
			return
		}
	}
}

func TestRealPICompactionPromptHook(t *testing.T) {
	config := requireRealPITestPrereqs(t)
	config.compactionPrompt = "Summarize the conversation in concise markdown with sections Goal, Progress, and Next Steps."

	client, err := startRealPIClient(t, config)
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	runCtx, cancelRun := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancelRun()

	if _, err := client.Run(runCtx, sdk.PromptRequest{Message: "Reply with exactly: COMPACTION_HOOK_READY"}); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	compactCtx, cancelCompact := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancelCompact()

	result, err := client.Compact(compactCtx, "")
	if err != nil {
		t.Fatalf("Compact failed: %v", err)
	}
	if len(result.Details) == 0 || string(result.Details) == "null" {
		t.Fatal("expected compaction details from hook")
	}

	var details map[string]any
	if err := json.Unmarshal(result.Details, &details); err != nil {
		t.Fatalf("failed to decode compaction details: %v", err)
	}
	if details["source"] != "pi-golang-compaction-prompt" {
		t.Fatalf("expected compaction source marker, got %+v", details)
	}
	if strings.TrimSpace(fmt.Sprint(details["promptHash"])) == "" {
		t.Fatalf("expected promptHash marker in details, got %+v", details)
	}
}

func TestRealPIWrapperContracts(t *testing.T) {
	config := requireRealPITestPrereqs(t)

	client, err := startRealPIClient(t, config)
	if err != nil {
		t.Fatalf("StartOneShot failed: %v", err)
	}
	defer client.Close()

	if _, err := client.GetState(nil); !errors.Is(err, sdk.ErrNilContext) {
		t.Fatalf("expected ErrNilContext from GetState(nil), got %v", err)
	}
	if err := client.Abort(nil); !errors.Is(err, sdk.ErrNilContext) {
		t.Fatalf("expected ErrNilContext from Abort(nil), got %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	state, err := client.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if strings.TrimSpace(state.SessionID) == "" {
		t.Fatal("expected session id")
	}
	if state.ContextWindow <= 0 {
		t.Fatalf("expected context window > 0, got %d", state.ContextWindow)
	}
}

func startRealPIClient(t *testing.T, config realPIConfig) (*sdk.OneShotClient, error) {
	t.Helper()
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(t.TempDir(), "pi-agent"))

	opts := sdk.DefaultOneShotOptions()
	opts.AppName = "pi-golang-realpi"
	opts.Mode = config.mode
	opts.Auth = config.auth
	opts.CompactionPrompt = config.compactionPrompt
	opts.InheritEnvironment = false
	opts.SeedAuthFromHome = false
	if path := strings.TrimSpace(os.Getenv("PATH")); path != "" {
		opts.Environment = map[string]string{"PATH": path}
	}
	return sdk.StartOneShot(opts)
}

func requireRealPITestPrereqs(t *testing.T) realPIConfig {
	t.Helper()
	required := os.Getenv("PI_REAL_REQUIRED") == "1"

	if os.Getenv("PI_REAL") != "1" {
		if required {
			t.Fatal("PI_REAL_REQUIRED=1 requires PI_REAL=1")
		}
		t.Skip("set PI_REAL=1 to run real pi integration tests")
	}
	if _, err := exec.LookPath("pi"); err != nil {
		if required {
			t.Fatalf("PI_REAL_REQUIRED=1 but pi CLI not found in PATH: %v", err)
		}
		t.Skip("pi CLI not found in PATH")
	}

	auth, invalidFiles := discoverProviderAuth()
	if invalidFiles != "" {
		if required {
			t.Fatalf("PI_REAL_REQUIRED=1 but credential file paths are invalid: %s", invalidFiles)
		}
		t.Skipf("credential file paths invalid: %s", invalidFiles)
	}
	mode, ok := preferredIntegrationMode(auth)
	if !ok {
		if required {
			t.Fatal("PI_REAL_REQUIRED=1 but no supported integration auth found (set ANTHROPIC_API_KEY[/_FILE] or OPENAI_API_KEY[/_FILE])")
		}
		t.Skip("set ANTHROPIC_API_KEY[/_FILE] or OPENAI_API_KEY[/_FILE]")
	}

	return realPIConfig{auth: auth, mode: mode}
}

func preferredIntegrationMode(auth sdk.ProviderAuth) (sdk.Mode, bool) {
	if hasAnthropicAuth(auth) {
		return sdk.ModeSmart, true
	}
	if hasOpenAIAuth(auth) {
		return sdk.ModeCoding, true
	}
	return "", false
}

func hasAnthropicAuth(auth sdk.ProviderAuth) bool {
	return hasCredential(auth.Anthropic.APIKey) || hasCredential(auth.Anthropic.OAuthToken) || strings.TrimSpace(auth.Anthropic.TokenFilePath) != ""
}

func hasOpenAIAuth(auth sdk.ProviderAuth) bool {
	return hasCredential(auth.OpenAI.APIKey)
}

func hasCredential(credential sdk.Credential) bool {
	return strings.TrimSpace(credential.Value) != "" || strings.TrimSpace(credential.File) != ""
}

type authBinding struct {
	envKey string
	set    func(*sdk.ProviderAuth, sdk.Credential)
}

func discoverProviderAuth() (sdk.ProviderAuth, string) {
	auth := sdk.ProviderAuth{}
	invalid := []string{}

	bindings := []authBinding{
		{envKey: "ANTHROPIC_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Anthropic.APIKey = credential }},
		{envKey: "ANTHROPIC_OAUTH_TOKEN", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Anthropic.OAuthToken = credential }},
		{envKey: "OPENAI_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.OpenAI.APIKey = credential }},
		{envKey: "GEMINI_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Gemini.APIKey = credential }},
		{envKey: "MISTRAL_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Mistral.APIKey = credential }},
		{envKey: "GROQ_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Groq.APIKey = credential }},
		{envKey: "CEREBRAS_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Cerebras.APIKey = credential }},
		{envKey: "XAI_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.XAI.APIKey = credential }},
		{envKey: "OPENROUTER_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.OpenRouter.APIKey = credential }},
		{envKey: "ZAI_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.ZAI.APIKey = credential }},
		{envKey: "MINIMAX_API_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Minimax.APIKey = credential }},
		{envKey: "AWS_PROFILE", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Bedrock.Profile = credential }},
		{envKey: "AWS_ACCESS_KEY_ID", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Bedrock.AccessKeyID = credential }},
		{envKey: "AWS_SECRET_ACCESS_KEY", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Bedrock.SecretAccessKey = credential }},
		{envKey: "AWS_BEARER_TOKEN_BEDROCK", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Bedrock.BearerToken = credential }},
		{envKey: "AWS_REGION", set: func(auth *sdk.ProviderAuth, credential sdk.Credential) { auth.Bedrock.Region = credential }},
	}

	for _, binding := range bindings {
		credential, hasCredential, err := credentialFromEnv(binding.envKey)
		if err != nil {
			invalid = append(invalid, err.Error())
			continue
		}
		if hasCredential {
			binding.set(&auth, credential)
		}
	}

	if tokenFilePath := strings.TrimSpace(os.Getenv("ANTHROPIC_TOKEN_FILE")); tokenFilePath != "" {
		if !fileExistsNonEmpty(tokenFilePath) {
			invalid = append(invalid, fmt.Sprintf("ANTHROPIC_TOKEN_FILE=%s", tokenFilePath))
		} else {
			auth.Anthropic.TokenFilePath = tokenFilePath
		}
	}

	return auth, strings.Join(invalid, ", ")
}

func credentialFromEnv(envKey string) (sdk.Credential, bool, error) {
	value := strings.TrimSpace(os.Getenv(envKey))
	filePath := strings.TrimSpace(os.Getenv(envKey + "_FILE"))

	if value != "" && filePath != "" {
		return sdk.Credential{}, false, fmt.Errorf("%s and %s_FILE both set", envKey, envKey)
	}
	if value != "" {
		return sdk.Credential{Value: value}, true, nil
	}
	if filePath != "" {
		if !fileExistsNonEmpty(filePath) {
			return sdk.Credential{}, false, fmt.Errorf("%s_FILE=%s", envKey, filePath)
		}
		return sdk.Credential{File: filePath}, true, nil
	}
	return sdk.Credential{}, false, nil
}

func fileExistsNonEmpty(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return info.Size() > 0
}
