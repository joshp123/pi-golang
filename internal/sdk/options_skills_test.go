package sdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeOneShotOptionsDefaultsSkillsModeToDisabled(t *testing.T) {
	options := OneShotOptions{}
	normalized, err := normalizeOneShotOptions(options)
	if err != nil {
		t.Fatalf("normalizeOneShotOptions returned error: %v", err)
	}
	if normalized.Skills.Mode != SkillsModeDisabled {
		t.Fatalf("expected skills mode %q, got %q", SkillsModeDisabled, normalized.Skills.Mode)
	}
	if len(normalized.Skills.Paths) != 0 {
		t.Fatalf("expected no skill paths, got %v", normalized.Skills.Paths)
	}
}

func TestNormalizeOneShotOptionsDisabledRejectsPaths(t *testing.T) {
	options := DefaultOneShotOptions()
	options.Skills = SkillsOptions{Mode: SkillsModeDisabled, Paths: []string{"./skills"}}

	_, err := normalizeOneShotOptions(options)
	if err == nil {
		t.Fatal("expected error for disabled skills mode with paths")
	}
}

func TestNormalizeOneShotOptionsAmbientRejectsPaths(t *testing.T) {
	options := DefaultOneShotOptions()
	options.Skills = SkillsOptions{Mode: SkillsModeAmbient, Paths: []string{"./skills"}}

	_, err := normalizeOneShotOptions(options)
	if err == nil {
		t.Fatal("expected error for ambient skills mode with paths")
	}
}

func TestNormalizeOneShotOptionsExplicitRequiresPaths(t *testing.T) {
	options := DefaultOneShotOptions()
	options.Skills = SkillsOptions{Mode: SkillsModeExplicit}

	_, err := normalizeOneShotOptions(options)
	if err == nil {
		t.Fatal("expected error for explicit skills mode without paths")
	}
}

func TestNormalizeOneShotOptionsExplicitResolvesAndDedupesPaths(t *testing.T) {
	workDir := t.TempDir()
	skillDir := filepath.Join(workDir, "frontend")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	options := DefaultOneShotOptions()
	options.WorkDir = workDir
	options.Skills = SkillsOptions{Mode: SkillsModeExplicit, Paths: []string{"./frontend", "./frontend", "  ./frontend  "}}

	normalized, err := normalizeOneShotOptions(options)
	if err != nil {
		t.Fatalf("normalizeOneShotOptions returned error: %v", err)
	}

	if normalized.Skills.Mode != SkillsModeExplicit {
		t.Fatalf("expected mode %q, got %q", SkillsModeExplicit, normalized.Skills.Mode)
	}
	if len(normalized.Skills.Paths) != 1 {
		t.Fatalf("expected one normalized path, got %v", normalized.Skills.Paths)
	}
	expected := filepath.Clean(filepath.Join(workDir, "frontend"))
	if normalized.Skills.Paths[0] != expected {
		t.Fatalf("expected normalized path %q, got %q", expected, normalized.Skills.Paths[0])
	}
}

func TestNormalizeOneShotOptionsExplicitAcceptsMarkdownFile(t *testing.T) {
	workDir := t.TempDir()
	skillFile := filepath.Join(workDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# skill"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	options := DefaultOneShotOptions()
	options.WorkDir = workDir
	options.Skills = SkillsOptions{Mode: SkillsModeExplicit, Paths: []string{"./SKILL.md"}}

	normalized, err := normalizeOneShotOptions(options)
	if err != nil {
		t.Fatalf("normalizeOneShotOptions returned error: %v", err)
	}
	if len(normalized.Skills.Paths) != 1 || normalized.Skills.Paths[0] != skillFile {
		t.Fatalf("unexpected normalized skill paths: %v", normalized.Skills.Paths)
	}
}
