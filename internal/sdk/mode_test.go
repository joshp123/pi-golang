package sdk

import "testing"

func TestResolveModelConfigSmart(t *testing.T) {
	cfg, err := resolveModelConfig(ModeSmart, DragonsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, DefaultProvider, DefaultModel, DefaultThinking)
}

func TestResolveModelConfigDumb(t *testing.T) {
	cfg, err := resolveModelConfig(ModeDumb, DragonsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, DefaultProvider, DefaultModel, DefaultDumbThinking)
}

func TestResolveModelConfigFast(t *testing.T) {
	cfg, err := resolveModelConfig(ModeFast, DragonsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, DefaultProvider, DefaultFastModel, DefaultDumbThinking)
}

func TestResolveModelConfigCoding(t *testing.T) {
	cfg, err := resolveModelConfig(ModeCoding, DragonsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, DefaultCodingProvider, DefaultCodingModel, DefaultCodingThinking)
}

func TestResolveModelConfigDragons(t *testing.T) {
	cfg, err := resolveModelConfig(ModeDragons, DragonsOptions{
		Provider: "anthropic",
		Model:    "claude-opus-4-5",
		Thinking: "high",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, "anthropic", "claude-opus-4-5", "high")
}

func TestResolveModelConfigDragonsValidation(t *testing.T) {
	if _, err := resolveModelConfig(ModeDragons, DragonsOptions{}); err == nil {
		t.Fatalf("expected error for missing dragons values")
	}
}

func TestResolveModelConfigDragonsMisuse(t *testing.T) {
	_, err := resolveModelConfig(ModeSmart, DragonsOptions{
		Provider: "anthropic",
		Model:    "claude-opus-4-5",
		Thinking: "high",
	})
	if err == nil {
		t.Fatalf("expected error for dragons values without dragons mode")
	}
}

func TestResolveModelConfigInvalidMode(t *testing.T) {
	if _, err := resolveModelConfig(Mode("nope"), DragonsOptions{}); err == nil {
		t.Fatalf("expected error for invalid mode")
	}
}

func TestDefaultOptionsEnvironmentPolicy(t *testing.T) {
	oneShot := DefaultOneShotOptions()
	session := DefaultSessionOptions()

	if oneShot.InheritEnvironment {
		t.Fatal("default one-shot options should not inherit host environment")
	}
	if session.InheritEnvironment {
		t.Fatal("default session options should not inherit host environment")
	}
	if !oneShot.SeedAuthFromHome {
		t.Fatal("default one-shot options should seed auth from ~/.pi/agent")
	}
	if !session.SeedAuthFromHome {
		t.Fatal("default session options should seed auth from ~/.pi/agent")
	}
	if oneShot.Skills.Mode != SkillsModeDisabled {
		t.Fatalf("default one-shot options should disable ambient skills, got %q", oneShot.Skills.Mode)
	}
	if session.Skills.Mode != SkillsModeDisabled {
		t.Fatalf("default session options should disable ambient skills, got %q", session.Skills.Mode)
	}
}

func assertConfig(t *testing.T, cfg modelConfig, provider string, model string, thinking string) {
	t.Helper()
	if cfg.provider != provider || cfg.model != model || cfg.thinking != thinking {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
