package sdk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joshp123/pi-golang/internal/testsupport"
)

func TestCreateManagedCompactionHookWritesBundle(t *testing.T) {
	hook, err := createManagedCompactionHook("summarize all diffs")
	if err != nil {
		t.Fatalf("createManagedCompactionHook returned error: %v", err)
	}
	defer hook.cleanup()

	if strings.TrimSpace(hook.promptHash) == "" {
		t.Fatal("expected prompt hash")
	}
	if _, err := os.Stat(hook.extensionPath); err != nil {
		t.Fatalf("expected extension file to exist: %v", err)
	}
	if _, err := os.Stat(hook.promptPath); err != nil {
		t.Fatalf("expected prompt file to exist: %v", err)
	}

	promptContent, err := os.ReadFile(hook.promptPath)
	if err != nil {
		t.Fatalf("failed to read prompt file: %v", err)
	}
	if string(promptContent) != "summarize all diffs" {
		t.Fatalf("unexpected prompt content: %q", string(promptContent))
	}

	extensionContent, err := os.ReadFile(hook.extensionPath)
	if err != nil {
		t.Fatalf("failed to read extension file: %v", err)
	}
	source := string(extensionContent)
	if !strings.Contains(source, compactionPromptFileEnv) {
		t.Fatalf("expected extension source to reference %s", compactionPromptFileEnv)
	}
	if !strings.Contains(source, compactionPromptHashEnv) {
		t.Fatalf("expected extension source to reference %s", compactionPromptHashEnv)
	}
}

func TestManagedCompactionHookInjectEnvironment(t *testing.T) {
	hook, err := createManagedCompactionHook("summarize all diffs")
	if err != nil {
		t.Fatalf("createManagedCompactionHook returned error: %v", err)
	}
	defer hook.cleanup()

	env := map[string]string{}
	hook.injectEnvironment(env)

	if env[compactionPromptFileEnv] != hook.promptPath {
		t.Fatalf("unexpected prompt file env value: %q", env[compactionPromptFileEnv])
	}
	if env[compactionPromptHashEnv] != hook.promptHash {
		t.Fatalf("unexpected prompt hash env value: %q", env[compactionPromptHashEnv])
	}
}

func TestStartClientCompactionPromptInjectsHookAndCleansUp(t *testing.T) {
	testsupport.SetupFakePI(t, "never_respond")

	client, err := startClient(startConfig{
		appName:          "pi-golang-test",
		mode:             ModeSmart,
		auth:             ProviderAuth{Anthropic: AnthropicAuth{APIKey: Credential{Value: "test-key"}}},
		compactionPrompt: "keep concise summary with next steps",
		useSession:       false,
	})
	if err != nil {
		t.Fatalf("startClient returned error: %v", err)
	}

	env := envSliceToMap(client.process.Env)
	promptPath := strings.TrimSpace(env[compactionPromptFileEnv])
	if promptPath == "" {
		_ = client.Close()
		t.Fatalf("expected %s env var", compactionPromptFileEnv)
	}
	if strings.TrimSpace(env[compactionPromptHashEnv]) == "" {
		_ = client.Close()
		t.Fatalf("expected %s env var", compactionPromptHashEnv)
	}
	if _, err := os.Stat(promptPath); err != nil {
		_ = client.Close()
		t.Fatalf("expected prompt file to exist: %v", err)
	}

	extensionPath := valueAfterFlag(client.process.Args, "--extension")
	if strings.TrimSpace(extensionPath) == "" {
		_ = client.Close()
		t.Fatalf("expected --extension argument, args=%v", client.process.Args)
	}
	if _, err := os.Stat(extensionPath); err != nil {
		_ = client.Close()
		t.Fatalf("expected extension file to exist: %v", err)
	}

	bundleDir := filepath.Dir(extensionPath)
	if err := client.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if _, err := os.Stat(bundleDir); !os.IsNotExist(err) {
		t.Fatalf("expected bundle dir to be removed, stat err=%v", err)
	}
}

func envSliceToMap(entries []string) map[string]string {
	values := map[string]string{}
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		values[key] = value
	}
	return values
}

func valueAfterFlag(args []string, flag string) string {
	for index := 0; index < len(args)-1; index++ {
		if args[index] == flag {
			return args[index+1]
		}
	}
	return ""
}
