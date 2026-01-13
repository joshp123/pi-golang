package pi

import "testing"

func TestResolveModelConfigSmart(t *testing.T) {
	cfg, err := resolveModelConfig(Options{Mode: ModeSmart})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, DefaultProvider, DefaultModel, DefaultThinking)
}

func TestResolveModelConfigDumb(t *testing.T) {
	cfg, err := resolveModelConfig(Options{Mode: ModeDumb})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, DefaultProvider, DefaultModel, DefaultDumbThinking)
}

func TestResolveModelConfigFast(t *testing.T) {
	cfg, err := resolveModelConfig(Options{Mode: ModeFast})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, DefaultProvider, DefaultFastModel, DefaultDumbThinking)
}

func TestResolveModelConfigCoding(t *testing.T) {
	cfg, err := resolveModelConfig(Options{Mode: ModeCoding})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, DefaultCodingProvider, DefaultCodingModel, DefaultCodingThinking)
}

func TestResolveModelConfigDragons(t *testing.T) {
	cfg, err := resolveModelConfig(Options{Mode: ModeDragons, Dragons: DragonsOptions{
		Provider: "anthropic",
		Model:    "claude-opus-4-5",
		Thinking: "high",
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertConfig(t, cfg, "anthropic", "claude-opus-4-5", "high")
}

func TestResolveModelConfigDragonsValidation(t *testing.T) {
	if _, err := resolveModelConfig(Options{Mode: ModeDragons}); err == nil {
		t.Fatalf("expected error for missing dragons values")
	}
}

func TestResolveModelConfigDragonsMisuse(t *testing.T) {
	_, err := resolveModelConfig(Options{Mode: ModeSmart, Dragons: DragonsOptions{
		Provider: "anthropic",
		Model:    "claude-opus-4-5",
		Thinking: "high",
	}})
	if err == nil {
		t.Fatalf("expected error for dragons values without dragons mode")
	}
}

func TestResolveModelConfigInvalidMode(t *testing.T) {
	if _, err := resolveModelConfig(Options{Mode: Mode("nope")}); err == nil {
		t.Fatalf("expected error for invalid mode")
	}
}

func assertConfig(t *testing.T, cfg modelConfig, provider string, model string, thinking string) {
	t.Helper()
	if cfg.provider != provider || cfg.model != model || cfg.thinking != thinking {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
