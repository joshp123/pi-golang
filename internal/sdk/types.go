package sdk

import "encoding/json"

type Event struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

const (
	EventTypeAgentEnd            = "agent_end"
	EventTypeMessageUpdate       = "message_update"
	EventTypeAutoCompactionStart = "auto_compaction_start"
	EventTypeAutoCompactionEnd   = "auto_compaction_end"
	EventTypeAutoRetryStart      = "auto_retry_start"
	EventTypeAutoRetryEnd        = "auto_retry_end"
	EventTypeProcessDied         = "process_died"
	EventTypeSubscriptionDrop    = "subscription_drop"
)

type SubscriptionMode string

const (
	SubscriptionModeDrop  SubscriptionMode = "drop"
	SubscriptionModeBlock SubscriptionMode = "block"
	SubscriptionModeRing  SubscriptionMode = "ring"
)

type SubscriptionPolicy struct {
	Buffer        int
	Mode          SubscriptionMode
	EmitDropEvent bool
}

func DefaultSubscriptionPolicy() SubscriptionPolicy {
	return SubscriptionPolicy{Buffer: 128, Mode: SubscriptionModeDrop}
}

type StreamingBehavior string

const (
	StreamingBehaviorSteer    StreamingBehavior = "steer"
	StreamingBehaviorFollowUp StreamingBehavior = "followUp"
)

type PromptRequest struct {
	Message           string
	Images            []ImageContent
	StreamingBehavior StreamingBehavior
}

type ImageContent struct {
	Data     string `json:"data"`
	MIMEType string `json:"mimeType"`
}

type Usage struct {
	Input       int   `json:"input"`
	Output      int   `json:"output"`
	CacheRead   int   `json:"cacheRead"`
	CacheWrite  int   `json:"cacheWrite"`
	TotalTokens int   `json:"totalTokens,omitempty"`
	Cost        *Cost `json:"cost,omitempty"`
}

type Cost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
	Total      float64 `json:"total"`
}

type RunResult struct {
	Text  string
	Usage *Usage
}

type TerminalStatus string

const (
	TerminalStatusCompleted TerminalStatus = "completed"
	TerminalStatusFailed    TerminalStatus = "failed"
	TerminalStatusAborted   TerminalStatus = "aborted"
)

type TerminalReason string

type TerminalOutcome struct {
	Status         TerminalStatus
	Text           string
	StopReason     string
	TerminalReason TerminalReason
	ErrorMessage   string
	Usage          *Usage
}

type RunDetailedResult struct {
	Outcome             TerminalOutcome
	AutoCompactionStart *AutoCompactionStartEvent
	AutoCompactionEnd   *AutoCompactionEndEvent
	AutoRetryStart      *AutoRetryStartEvent
	AutoRetryEnd        *AutoRetryEndEvent
}

type CompletionClass string

const (
	CompletionClassOK              CompletionClass = "ok"
	CompletionClassOKAfterRecovery CompletionClass = "ok_after_recovery"
	CompletionClassAborted         CompletionClass = "aborted"
	CompletionClassFailed          CompletionClass = "failed"
)

type RecoveryFacts struct {
	CompactionObserved bool
	OverflowDetected   bool
	Recovered          bool
}

type ManagedSummary struct {
	Class CompletionClass
	Facts RecoveryFacts
}

type BrokenCause string

const (
	BrokenCauseProcessDied BrokenCause = "process_died"
	BrokenCauseProtocol    BrokenCause = "protocol_violation"
	BrokenCauseClient      BrokenCause = "client_runtime"
)

type ShareResult struct {
	GistURL    string
	GistID     string
	PreviewURL string
}

type ModelInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name,omitempty"`
	API           string   `json:"api,omitempty"`
	Provider      string   `json:"provider"`
	BaseURL       string   `json:"baseUrl,omitempty"`
	Reasoning     bool     `json:"reasoning"`
	Input         []string `json:"input,omitempty"`
	ContextWindow int      `json:"contextWindow"`
	MaxTokens     int      `json:"maxTokens"`
	Cost          *Cost    `json:"cost,omitempty"`
}

type SessionState struct {
	Model                 *ModelInfo `json:"model,omitempty"`
	ThinkingLevel         string     `json:"thinkingLevel,omitempty"`
	IsStreaming           bool       `json:"isStreaming"`
	IsCompacting          bool       `json:"isCompacting"`
	SteeringMode          string     `json:"steeringMode,omitempty"`
	FollowUpMode          string     `json:"followUpMode,omitempty"`
	SessionID             string     `json:"sessionId"`
	SessionFile           string     `json:"sessionFile,omitempty"`
	SessionName           string     `json:"sessionName,omitempty"`
	AutoCompactionEnabled bool       `json:"autoCompactionEnabled"`
	MessageCount          int        `json:"messageCount"`
	PendingMessageCount   int        `json:"pendingMessageCount"`
	ContextWindow         int        `json:"-"`
}

type CompactResult struct {
	Summary          string          `json:"summary"`
	FirstKeptEntryID string          `json:"firstKeptEntryId"`
	TokensBefore     int             `json:"tokensBefore"`
	Details          json.RawMessage `json:"details,omitempty"`
}

type SkillLocation string

const (
	SkillLocationUser    SkillLocation = "user"
	SkillLocationProject SkillLocation = "project"
	SkillLocationPath    SkillLocation = "path"
	SkillLocationUnknown SkillLocation = "unknown"
)

type LoadedSkill struct {
	Name        string
	Description string
	Path        string
	Location    SkillLocation
}

type AgentMessage struct {
	Role              string          `json:"role"`
	Content           json.RawMessage `json:"content"`
	Usage             *Usage          `json:"usage,omitempty"`
	StopReason        string          `json:"stopReason,omitempty"`
	TerminalReason    TerminalReason  `json:"terminalReason,omitempty"`
	TerminalReasonAlt TerminalReason  `json:"terminal_reason,omitempty"`
	ErrorMessage      string          `json:"errorMessage,omitempty"`
}

type AgentEndEvent struct {
	Messages []AgentMessage `json:"messages"`
}

type AssistantMessageDelta struct {
	Type         string          `json:"type"`
	ContentIndex int             `json:"contentIndex,omitempty"`
	Delta        string          `json:"delta,omitempty"`
	Reason       string          `json:"reason,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
	Partial      json.RawMessage `json:"partial,omitempty"`
	Raw          json.RawMessage `json:"-"`
}

type MessageUpdateEvent struct {
	Message               AgentMessage          `json:"message"`
	AssistantMessageEvent AssistantMessageDelta `json:"assistantMessageEvent"`
}

type AutoCompactionStartEvent struct {
	Reason string `json:"reason"`
}

type AutoCompactionEndEvent struct {
	Result       *CompactResult `json:"result"`
	Aborted      bool           `json:"aborted"`
	WillRetry    bool           `json:"willRetry"`
	ErrorMessage string         `json:"errorMessage,omitempty"`
}

type AutoRetryStartEvent struct {
	Attempt      int    `json:"attempt"`
	MaxAttempts  int    `json:"maxAttempts"`
	DelayMS      int    `json:"delayMs"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type AutoRetryEndEvent struct {
	Success    bool   `json:"success"`
	Attempt    int    `json:"attempt"`
	FinalError string `json:"finalError,omitempty"`
}
