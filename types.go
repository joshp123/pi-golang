package pi

import "github.com/joshp123/pi-golang/internal/sdk"

type Event = sdk.Event

const (
	EventTypeAgentEnd            = sdk.EventTypeAgentEnd
	EventTypeMessageUpdate       = sdk.EventTypeMessageUpdate
	EventTypeAutoCompactionStart = sdk.EventTypeAutoCompactionStart
	EventTypeAutoCompactionEnd   = sdk.EventTypeAutoCompactionEnd
	EventTypeAutoRetryStart      = sdk.EventTypeAutoRetryStart
	EventTypeAutoRetryEnd        = sdk.EventTypeAutoRetryEnd
	EventTypeProcessDied         = sdk.EventTypeProcessDied
	EventTypeSubscriptionDrop    = sdk.EventTypeSubscriptionDrop
)

type SubscriptionMode = sdk.SubscriptionMode

const (
	SubscriptionModeDrop  = sdk.SubscriptionModeDrop
	SubscriptionModeBlock = sdk.SubscriptionModeBlock
	SubscriptionModeRing  = sdk.SubscriptionModeRing
)

type SubscriptionPolicy = sdk.SubscriptionPolicy

func DefaultSubscriptionPolicy() SubscriptionPolicy {
	return sdk.DefaultSubscriptionPolicy()
}

type StreamingBehavior = sdk.StreamingBehavior

const (
	StreamingBehaviorSteer    = sdk.StreamingBehaviorSteer
	StreamingBehaviorFollowUp = sdk.StreamingBehaviorFollowUp
)

type PromptRequest = sdk.PromptRequest
type ImageContent = sdk.ImageContent
type Usage = sdk.Usage
type Cost = sdk.Cost
type RunResult = sdk.RunResult

type TerminalStatus = sdk.TerminalStatus

const (
	TerminalStatusCompleted = sdk.TerminalStatusCompleted
	TerminalStatusFailed    = sdk.TerminalStatusFailed
	TerminalStatusAborted   = sdk.TerminalStatusAborted
)

type TerminalOutcome = sdk.TerminalOutcome
type RunDetailedResult = sdk.RunDetailedResult
type ShareResult = sdk.ShareResult
type ModelInfo = sdk.ModelInfo
type SessionState = sdk.SessionState
type CompactResult = sdk.CompactResult
type AgentMessage = sdk.AgentMessage
type AgentEndEvent = sdk.AgentEndEvent
type AssistantMessageDelta = sdk.AssistantMessageDelta
type MessageUpdateEvent = sdk.MessageUpdateEvent
type AutoCompactionStartEvent = sdk.AutoCompactionStartEvent
type AutoCompactionEndEvent = sdk.AutoCompactionEndEvent
type AutoRetryStartEvent = sdk.AutoRetryStartEvent
type AutoRetryEndEvent = sdk.AutoRetryEndEvent
