package sdk

import (
	"fmt"
	"os"
	"testing"

	"github.com/joshp123/pi-golang/internal/testsupport"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_PI_HELPER") != "1" {
		return
	}

	scenario := testsupport.ScenarioFromArgs(os.Args, "happy")
	if err := testsupport.RunScenario(scenario, os.Args, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "helper scenario failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
