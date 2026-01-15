package pi

import (
	"os"
	"strings"
	"testing"
)

func TestBuildEnvRespectsExistingPICodingAgentDir(t *testing.T) {
	// Save and restore original env
	original := os.Getenv("PI_CODING_AGENT_DIR")
	defer func() {
		if original != "" {
			os.Setenv("PI_CODING_AGENT_DIR", original)
		} else {
			os.Unsetenv("PI_CODING_AGENT_DIR")
		}
	}()

	customPath := "/custom/agent/dir"
	os.Setenv("PI_CODING_AGENT_DIR", customPath)

	env, err := buildEnv("test-app")
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}

	// Find PI_CODING_AGENT_DIR in the result
	var found string
	for _, e := range env {
		if strings.HasPrefix(e, "PI_CODING_AGENT_DIR=") {
			found = strings.TrimPrefix(e, "PI_CODING_AGENT_DIR=")
			break
		}
	}

	if found != customPath {
		t.Errorf("PI_CODING_AGENT_DIR not respected: got %q, want %q", found, customPath)
	}
}

func TestBuildEnvSetsDefaultPICodingAgentDir(t *testing.T) {
	// Save and restore original env
	original := os.Getenv("PI_CODING_AGENT_DIR")
	defer func() {
		if original != "" {
			os.Setenv("PI_CODING_AGENT_DIR", original)
		} else {
			os.Unsetenv("PI_CODING_AGENT_DIR")
		}
	}()

	os.Unsetenv("PI_CODING_AGENT_DIR")

	env, err := buildEnv("test-app")
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}

	// Find PI_CODING_AGENT_DIR in the result
	var found string
	for _, e := range env {
		if strings.HasPrefix(e, "PI_CODING_AGENT_DIR=") {
			found = strings.TrimPrefix(e, "PI_CODING_AGENT_DIR=")
			break
		}
	}

	if found == "" {
		t.Error("PI_CODING_AGENT_DIR not set when env var is unset")
	}

	// Should be under home directory with app name
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(found, home) {
		t.Errorf("PI_CODING_AGENT_DIR should be under home dir, got %q", found)
	}
}

func TestBuildEnvPassesAllowlistedVars(t *testing.T) {
	env, err := buildEnv("test-app")
	if err != nil {
		t.Fatalf("buildEnv failed: %v", err)
	}

	// HOME and PATH should always be present
	hasHome := false
	hasPath := false
	for _, e := range env {
		if strings.HasPrefix(e, "HOME=") {
			hasHome = true
		}
		if strings.HasPrefix(e, "PATH=") {
			hasPath = true
		}
	}

	if !hasHome {
		t.Error("HOME not in env")
	}
	if !hasPath {
		t.Error("PATH not in env")
	}
}
