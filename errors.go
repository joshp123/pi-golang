package pi

import "github.com/joshp123/pi-golang/internal/sdk"

var (
	ErrProcessDied               = sdk.ErrProcessDied
	ErrClientClosed              = sdk.ErrClientClosed
	ErrProtocolViolation         = sdk.ErrProtocolViolation
	ErrNilContext                = sdk.ErrNilContext
	ErrRunInProgress             = sdk.ErrRunInProgress
	ErrInvalidSubscriptionPolicy = sdk.ErrInvalidSubscriptionPolicy
)

type RPCError = sdk.RPCError
type MissingProviderAuthError = sdk.MissingProviderAuthError
