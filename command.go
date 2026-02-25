package pi

import "github.com/joshp123/pi-golang/internal/sdk"

type Command = sdk.Command

func ResolveCommand() (Command, error) {
	return sdk.ResolveCommand()
}
