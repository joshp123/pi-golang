package pi

import "github.com/joshp123/pi-golang/internal/sdk"

const (
	DefaultProvider       = sdk.DefaultProvider
	DefaultModel          = sdk.DefaultModel
	DefaultThinking       = sdk.DefaultThinking
	DefaultDumbThinking   = sdk.DefaultDumbThinking
	DefaultFastModel      = sdk.DefaultFastModel
	DefaultCodingProvider = sdk.DefaultCodingProvider
	DefaultCodingModel    = sdk.DefaultCodingModel
	DefaultCodingThinking = sdk.DefaultCodingThinking
)

type Mode = sdk.Mode

const (
	ModeSmart   = sdk.ModeSmart
	ModeDumb    = sdk.ModeDumb
	ModeFast    = sdk.ModeFast
	ModeCoding  = sdk.ModeCoding
	ModeDragons = sdk.ModeDragons
)

type DragonsOptions = sdk.DragonsOptions
type Credential = sdk.Credential
type APIKeyAuth = sdk.APIKeyAuth
type AnthropicAuth = sdk.AnthropicAuth
type BedrockAuth = sdk.BedrockAuth
type ProviderAuth = sdk.ProviderAuth
type SessionOptions = sdk.SessionOptions
type OneShotOptions = sdk.OneShotOptions

func DefaultSessionOptions() SessionOptions {
	return sdk.DefaultSessionOptions()
}

func DefaultOneShotOptions() OneShotOptions {
	return sdk.DefaultOneShotOptions()
}
