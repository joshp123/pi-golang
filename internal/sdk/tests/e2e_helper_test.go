package sdk_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	sdk "github.com/joshp123/pi-golang/internal/sdk"
	"github.com/joshp123/pi-golang/internal/testsupport"
)

func setupFakePI(t *testing.T, scenario string) {
	testsupport.SetupFakePI(t, scenario)
}

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

func readEventOrFail(t *testing.T, events <-chan sdk.Event) sdk.Event {
	t.Helper()
	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed")
		}
		return event
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
		return sdk.Event{}
	}
}
