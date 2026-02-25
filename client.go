package pi

import (
	"os"

	"github.com/joshp123/pi-golang/internal/sdk"
)

// Debug enables verbose logging when PI_DEBUG=1.
var Debug = os.Getenv("PI_DEBUG") == "1"

func init() {
	sdk.SetDebugEnabledProvider(func() bool { return Debug })
}

type Client = sdk.Client
type SessionClient = sdk.SessionClient
type OneShotClient = sdk.OneShotClient

func StartSession(options SessionOptions) (*SessionClient, error) {
	sdk.DefaultEnvAllowlist = DefaultEnvAllowlist
	sdk.DefaultEnvAllowPrefixes = DefaultEnvAllowPrefixes
	return sdk.StartSession(options)
}

func StartOneShot(options OneShotOptions) (*OneShotClient, error) {
	sdk.DefaultEnvAllowlist = DefaultEnvAllowlist
	sdk.DefaultEnvAllowPrefixes = DefaultEnvAllowPrefixes
	return sdk.StartOneShot(options)
}
