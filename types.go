package pi

import "encoding/json"

type rpcCommand map[string]any

type rpcResponse struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Command string          `json:"command,omitempty"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type Event struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

const (
	// Canonical RPC command types (upstream docs/rpc.md).
	rpcCommandPrompt     = "prompt"
	rpcCommandSteer      = "steer"
	rpcCommandFollowUp   = "follow_up"
	rpcCommandAbort      = "abort"
	rpcCommandGetState   = "get_state"
	rpcCommandNewSession = "new_session"
	rpcCommandCompact    = "compact"
	rpcCommandExportHTML = "export_html"
)

const (
	eventTypeResponse           = "response"
	EventTypeAgentEnd           = "agent_end"
	EventTypeMessageUpdate      = "message_update"
	EventTypeAutoCompactionEnd  = "auto_compaction_end"
	EventTypeProcessDied        = "process_died"
	EventTypeSubscriptionDrop   = "subscription_drop"
	eventTypeParseError         = "parse_error"
	eventTypeResponseParseError = "response_parse_error"
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

type AgentMessage struct {
	Role         string          `json:"role"`
	Content      json.RawMessage `json:"content"`
	Usage        *Usage          `json:"usage,omitempty"`
	StopReason   string          `json:"stopReason,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
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

type AutoCompactionEndEvent struct {
	Result       *CompactResult `json:"result"`
	Aborted      bool           `json:"aborted"`
	WillRetry    bool           `json:"willRetry"`
	ErrorMessage string         `json:"errorMessage,omitempty"`
}
