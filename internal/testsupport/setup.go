package testsupport

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func SetupFakePI(t *testing.T, scenario string) {
	t.Helper()

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable failed: %v", err)
	}

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "pi")
	script := fmt.Sprintf("#!/usr/bin/env bash\nset -euo pipefail\nGO_WANT_PI_HELPER=1 exec %q -test.run '^TestHelperProcess$' -- %q \"$@\"\n", exe, scenario)
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake pi: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(t.TempDir(), "pi-agent"))
}

func ScenarioFromArgs(args []string, fallback string) string {
	scenario := fallback
	for index, arg := range args {
		if arg == "--" && index+1 < len(args) {
			return args[index+1]
		}
	}
	return scenario
}
