package sdk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joshp123/pi-golang/internal/testsupport"
)

func TestStartClientDefaultDisablesSkillDiscovery(t *testing.T) {
	testsupport.SetupFakePI(t, "never_respond")

	client, err := startClient(startConfig{
		appName:    "pi-golang-test",
		mode:       ModeSmart,
		auth:       ProviderAuth{Anthropic: AnthropicAuth{APIKey: Credential{Value: "test-key"}}},
		useSession: false,
	})
	if err != nil {
		t.Fatalf("startClient returned error: %v", err)
	}
	defer client.Close()

	if !hasFlag(client.process.Args, "--no-skills") {
		t.Fatalf("expected --no-skills in args, got %v", client.process.Args)
	}
}

func TestStartClientExplicitSkillsAddsSkillArgs(t *testing.T) {
	testsupport.SetupFakePI(t, "happy")

	skillDir := filepath.Join(t.TempDir(), "frontend")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	client, err := startClient(startConfig{
		appName: "pi-golang-test",
		mode:    ModeSmart,
		auth:    ProviderAuth{Anthropic: AnthropicAuth{APIKey: Credential{Value: "test-key"}}},
		skills: SkillsOptions{
			Mode:  SkillsModeExplicit,
			Paths: []string{skillDir},
		},
		useSession: false,
	})
	if err != nil {
		t.Fatalf("startClient returned error: %v", err)
	}
	defer client.Close()

	if !hasFlag(client.process.Args, "--no-skills") {
		t.Fatalf("expected --no-skills in args, got %v", client.process.Args)
	}
	if valueAfterFlag(client.process.Args, "--skill") != skillDir {
		t.Fatalf("expected --skill %q, got args=%v", skillDir, client.process.Args)
	}
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}
