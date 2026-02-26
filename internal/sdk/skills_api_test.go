package sdk

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joshp123/pi-golang/internal/testsupport"
)

func TestListLoadedSkillsReturnsExplicitSkills(t *testing.T) {
	testsupport.SetupFakePI(t, "happy")

	skillDir := filepath.Join(t.TempDir(), "frontend")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	options := DefaultOneShotOptions()
	options.Auth.Anthropic.APIKey = Credential{Value: "test-key"}
	options.Skills = SkillsOptions{Mode: SkillsModeExplicit, Paths: []string{skillDir}}

	client, err := StartOneShot(options)
	if err != nil {
		t.Fatalf("StartOneShot returned error: %v", err)
	}
	defer client.Close()

	skills, err := client.ListLoadedSkills(context.Background())
	if err != nil {
		t.Fatalf("ListLoadedSkills returned error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 loaded skill, got %d (%+v)", len(skills), skills)
	}
	if skills[0].Path != skillDir {
		t.Fatalf("expected skill path %q, got %q", skillDir, skills[0].Path)
	}
	if skills[0].Location != SkillLocationPath {
		t.Fatalf("expected skill location %q, got %q", SkillLocationPath, skills[0].Location)
	}
}

func TestStartOneShotExplicitSkillsRejectsUnexpectedLoadedSkill(t *testing.T) {
	testsupport.SetupFakePI(t, "skills_unexpected")

	skillDir := filepath.Join(t.TempDir(), "frontend")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	options := DefaultOneShotOptions()
	options.Auth.Anthropic.APIKey = Credential{Value: "test-key"}
	options.Skills = SkillsOptions{Mode: SkillsModeExplicit, Paths: []string{skillDir}}

	_, err := StartOneShot(options)
	if err == nil {
		t.Fatal("expected StartOneShot to fail when ambient skill appears")
	}
	if !strings.Contains(err.Error(), "outside explicit skills paths") {
		t.Fatalf("expected explicit-skills verification error, got %v", err)
	}
}
