# pi-golang

Go SDK for the pi coding agent using a managed `pi --mode rpc` process.

Thanks to Mario Zechner and the pi project: [pi-mono](https://github.com/badlogic/pi-mono).

## Design goals

- One process per client (no per-call shelling)
- One obvious request shape (`PromptRequest`) for all user messages
- Layered architecture, tree mirrors runtime:
  - thin RPC mirror (typed methods, 1:1 with upstream commands)
  - transport/process/runtime plumbing
  - stream fanout/backpressure runtime
  - batteries and typed decoders
- Explicit failure semantics (`ErrProcessDied`, `ErrClientClosed`, `ErrRunInProgress`)
- Strict contracts over silent fallbacks

## Engineering principles encoded

- Ontology-first naming: one canonical term per concept, no core synonyms.
- User-mechanics-first: API behavior in 3 bullets, architecture follows those bullets.
- KISS/YAGNI: thin RPC mirror + thin runtime wrappers + batteries only for ergonomics.
- Discoverability-first: root façade + explicit implementation/runtime layers (`internal/sdk`, `internal/rpc`, `internal/transport`, `internal/stream`).
- Agentic self-debug support: deterministic lifecycle/errors and hermetic e2e fake harness.

## Quick start

```go
opts := pi.DefaultOneShotOptions()
opts.Auth.Anthropic.APIKey = pi.Credential{File: "/run/agenix/anthropic-api-key"}
opts.Environment = map[string]string{"PATH": os.Getenv("PATH")}

client, err := pi.StartOneShot(opts)
if err != nil {
    // handle
}
defer client.Close()

result, err := client.Run(context.Background(), pi.PromptRequest{
    Message: "Summarize this in one sentence.",
})
if err != nil {
    // handle
}

fmt.Println(result.Text)
```

## User mechanics (canonical, 3 bullets)

- Start work: call `Prompt` (async) or `Run` (sync helper).
- Control in-flight work: call `Steer`, `FollowUp`, or `Abort`.
- Observe and terminate cleanly: consume `Subscribe` events, then `Close` when done.

The internal architecture follows this behavior directly: thin RPC mirror for commands, batteries only for ergonomics.

## API layers

### Thin RPC mirror (upstream parity)

- `Prompt(ctx, PromptRequest)`
- `Steer(ctx, PromptRequest)`
- `FollowUp(ctx, PromptRequest)`
- `Abort(ctx)`
- `GetState(ctx)`
- `NewSession(ctx, parentSession)`
- `Compact(ctx, instructions)`
- `ExportHTML(ctx, outputPath)` (session client)

These stay close to upstream `docs/rpc.md` command contracts.

### Batteries (ergonomics)

- `Run(ctx, PromptRequest)` (sync helper waiting for `agent_end`)
- `Subscribe(SubscriptionPolicy)` (fanout/backpressure policy)
- Typed event decoders (`DecodeAgentEnd`, `DecodeMessageUpdate`, `DecodeAutoCompactionStart`, `DecodeAutoCompactionEnd`, `DecodeAutoRetryStart`, `DecodeAutoRetryEnd`, `DecodeTerminalOutcome`)
- Pure managed classifiers (`ClassifyManaged`, `ClassifyRunError`)
- `ShareSession(ctx)` (export + gist helper)

## Package / file map (ontology-first)

```text
.
├── client.go                 # public façade: constructors + client type exports
├── types.go                  # public ontology exports
├── errors.go                 # public error contracts
├── mode.go                   # public mode/options exports
├── command.go                # public command resolver export
├── env.go                    # public env allowlist exports
├── decode.go                 # public typed decoder exports
├── managed.go                # public batteries classifiers exports
└── internal/
    ├── sdk/                  # canonical implementation package
    │   ├── api_rpc.go        # thin RPC mirror methods
    │   ├── rpc_contracts.go  # RPC command contracts + response decoding
    │   ├── send.go           # request transport path
    │   ├── process.go        # stdout/stderr/process lifecycle
    │   ├── subscriptions.go  # subscription policy + stream adapters
    │   ├── batteries_*.go    # convenience helpers
    │   ├── decode.go         # typed event/terminal decoders
    │   ├── tests/            # user-facing black-box/e2e tests
    │   └── *_test.go         # white-box runtime invariants
    ├── rpc/                  # private wire terms/constants
    ├── transport/            # queue + pending request wrappers
    ├── stream/               # generic fanout/backpressure hub
    ├── runtime/              # generic primitives
    └── testsupport/          # fake pi harness for e2e
```

This encodes: root package = stable API façade; `internal/sdk` = behavior implementation; lower internal packages = explicit plumbing layers.

## Intent map (ontology-first)

- Ask: `Prompt`, `Run`
- Steer queued work: `Steer`, `FollowUp`
- Abort current work: `Abort`
- Inspect runtime/session: `GetState`, `Stderr`
- Compact/session mgmt: `Compact`, `NewSession`, `ExportHTML`, `ShareSession`
- Observe stream: `Subscribe` + typed event decoders
- Classify managed outcomes: `ClassifyManaged`, `ClassifyRunError`
- Lifecycle: `Close`

## One way to send messages

All message APIs use the same shape:

```go
type PromptRequest struct {
    Message           string
    Images            []pi.ImageContent
    StreamingBehavior pi.StreamingBehavior // only for prompt/run
}
```

### Prompt (async)

```go
err := client.Prompt(ctx, pi.PromptRequest{Message: "Analyze this file"})
```

### Prompt with streaming behavior (when already streaming)

```go
err := client.Prompt(ctx, pi.PromptRequest{
    Message:           "Switch strategy",
    StreamingBehavior: pi.StreamingBehaviorSteer,
})
```

### Queue controls

```go
err = client.Steer(ctx, pi.PromptRequest{Message: "Interrupt and do X"})
err = client.FollowUp(ctx, pi.PromptRequest{Message: "After that, do Y"})
```

### Abort current run

```go
err = client.Abort(ctx)
```

### Run (sync helper)

`Run` sends a prompt and waits for `agent_end`.

```go
res, err := client.Run(ctx, pi.PromptRequest{Message: "Explain the diff"})
```

### RunDetailed (sync helper + compaction/retry signals)

```go
detailed, err := client.RunDetailed(ctx, pi.PromptRequest{Message: "Explain the diff"})
// detailed.Outcome
// detailed.AutoCompactionStart / detailed.AutoCompactionEnd
// detailed.AutoRetryStart / detailed.AutoRetryEnd
```

### Managed classification helpers (pure functions)

```go
summary := pi.ClassifyManaged(detailed)
// summary.Class: ok | ok_after_recovery | aborted | failed
// summary.Facts.CompactionObserved
// summary.Facts.OverflowDetected
// summary.Facts.Recovered

if runErr != nil {
    cause, broken := pi.ClassifyRunError(runErr)
    _ = cause
    _ = broken
}
```

## Environment + auth control (explicit)

```go
opts := pi.DefaultOneShotOptions() // defaults: InheritEnvironment=false, SeedAuthFromHome=true
opts.InheritEnvironment = false    // explicit env only
opts.SeedAuthFromHome = false      // optional: disable ~/.pi/agent auth seeding
opts.Auth.Anthropic.APIKey = pi.Credential{File: "/run/agenix/anthropic-api-key"}
opts.Environment = map[string]string{
    "PATH": os.Getenv("PATH"), // explicitly pass what you want
}

// Optional: custom prompt for compaction summaries (manual + auto compaction).
// Leave empty to use upstream default compaction behavior.
opts.CompactionPrompt = "Summarize code changes, decisions, and next steps for handoff."

client, err := pi.StartOneShot(opts)
```

## Event subscription

`Subscribe` is for callers that need live runtime events instead of only final
`Run` output.

Typical use-cases:
- streaming chat UIs (react to `message_update` as text arrives),
- audit/event logs (persist all event envelopes),
- custom orchestrators (watch compaction/retry/process lifecycle events),
- integrations that need raw event data for domain-specific parsing.

There is one subscription API:

```go
events, cancel, err := client.Subscribe(pi.SubscriptionPolicy{
    Buffer: 128,
    Mode:   pi.SubscriptionModeRing,
})
if err != nil {
    // handle
}
defer cancel()

_ = events
```

Backpressure modes:
- `drop`: drop new event when buffer full (lowest latency, lossy)
- `block`: wait for consumer (lossless, can stall producer path)
- `ring`: keep latest events by dropping oldest when full (UI-friendly)

Invalid policies fail with `ErrInvalidSubscriptionPolicy`.

## Typed runtime wrappers

```go
state, err := client.GetState(ctx)
cancelled, err := client.NewSession(ctx, "")
compacted, err := client.Compact(ctx, "Focus on code changes")
err = client.Abort(ctx)

_ = state
_ = cancelled
_ = compacted
```

## Typed event decoders

```go
switch event.Type {
case "agent_end":
    parsed, err := pi.DecodeAgentEnd(event.Raw)
    _ = parsed
    _ = err
case "message_update":
    parsed, err := pi.DecodeMessageUpdate(event.Raw)
    _ = parsed
    _ = err
case "auto_compaction_end":
    parsed, err := pi.DecodeAutoCompactionEnd(event.Raw)
    _ = parsed
    _ = err
}
```

Canonical terminal outcome (`agent_end` payload):

```go
outcome, err := pi.DecodeTerminalOutcome(event.Raw)
if err == nil {
    // outcome.Status: completed | failed | aborted
    // outcome.Text, outcome.StopReason, outcome.ErrorMessage, outcome.Usage
}
```

## API contract

- Contexts:
  - RPC methods require non-nil context (`ErrNilContext`).
  - If context has no deadline, a default 2m timeout is applied.
- Thin mirror methods (`Prompt`, `Steer`, `FollowUp`, `Abort`, `GetState`, `NewSession`, `Compact`, `ExportHTML`) map 1:1 to upstream RPC commands.
- Auth and environment are code-controlled via options:
  - `Auth ProviderAuth` carries explicit provider credentials (value or file path per field).
  - selected provider is presence-validated at startup (presence only, not credential validity).
  - credential env vars cannot be injected through `Environment`; use `Auth` only.
  - `Environment map[string]string` is for non-credential child env values (e.g. `PATH`).
  - `InheritEnvironment` controls allowlisted host env inheritance (default: `false`).
  - `SeedAuthFromHome` controls whether `~/.pi/agent/{auth.json,oauth.json}` are seeded into SDK agent dir (default: `true`).
  - `CompactionPrompt` (optional) installs an SDK-managed extension hook for manual/auto compaction and passes the prompt via file-backed env vars.
  - `PI_CODING_AGENT_DIR` is always set (explicit value wins; otherwise SDK-managed path).
- `GetState` guarantees `SessionState.ContextWindow > 0` (fallback from model metadata when needed; protocol violation otherwise).
- `Run` / `RunDetailed` are battery helpers:
  - single-flight per client (`ErrRunInProgress` on overlap)
  - send one `prompt`, wait for `agent_end`
  - on context cancellation while waiting, send best-effort `Abort` and return `ctx.Err()`
  - surface late async `prompt` failures (`response` frames) as `*RPCError`
- `RunDetailed` additionally returns typed compaction/retry signals from streamed events.
- `ClassifyManaged(RunDetailedResult)` is a pure classifier over typed run signals (`ok | ok_after_recovery | aborted | failed`) with no provider regex inference.
- `ClassifyRunError(error)` is a pure classifier for runtime/process breakage (`process_died`, `protocol_violation`, `client_runtime`) and keeps cancellation non-broken.
- `Abort(ctx)` sends upstream `{"type":"abort"}` and waits for command response.
- Process/lifecycle guarantees:
  - unexpected process exit fails pending requests with `ErrProcessDied`
  - emits exactly one `process_died` event
  - closes all subscriber channels after that event
  - `Close()` deterministically unblocks pending requests with `ErrClientClosed`
- Decoder strictness: RPC/event payloads must include explicit `type` values matching the expected envelope; missing/mismatched types fail fast.
- Overflow note: upstream RPC does not currently expose a typed `context_exhausted` reason. SDK exposes canonical terminal fields (`Status`, `StopReason`, `ErrorMessage`) plus typed compaction/retry events (`auto_compaction_*`, `auto_retry_*`) without provider-regex duplication.
- Raw transport path is internal (`send` + `internal/rpc.Command`/`internal/rpc.Response` are not public API).

## Modes

- `smart`: Opus 4.5 + high thinking
- `dumb`: Opus 4.5 + low thinking
- `fast`: Haiku 4.5 + low thinking
- `coding`: GPT-5.2 Codex + high thinking
- `dragons`: explicit `provider/model/thinking`

## Share session

Session clients can export + share via gist:

```go
result, err := client.ShareSession(context.Background())
fmt.Println(result.GistURL)
```

## Pre-commit checks

```bash
./scripts/check.sh
```

Runs:
- `go fix ./...`
- `gofmt -w .`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `staticcheck ./...` (if installed)
- smoke example (`go run ./examples/basic`)

## Real pi integration tests

These tests hit the actual `pi` binary (not fake harness scenarios).

```bash
PI_REAL=1 go test -tags=integration ./internal/sdk/tests -run TestRealPI -v
```

Release gate (mandatory real-pi run):

```bash
PI_REAL=1 ./scripts/check-release.sh
```

`check-release.sh` runs integration tests with `PI_REAL_REQUIRED=1`, so missing
credentials or missing `pi` binary fail the gate.

Prereqs:
- `pi` on PATH
- explicit credentials provided via env (`ANTHROPIC_API_KEY=...`, `ANTHROPIC_OAUTH_TOKEN=...`) or file-path env (`ANTHROPIC_API_KEY_FILE=/run/agenix/anthropic-api-key`)

Local dev (recommended): explicitly export a real token from `~/.pi/agent/auth.json`, then run gate:

```bash
ANTHROPIC_OAUTH_TOKEN="$(python -c 'import json, pathlib; p=pathlib.Path("~/.pi/agent/auth.json").expanduser(); print(json.loads(p.read_text()).get("anthropic", {}).get("access", ""), end="")')" \
PI_REAL=1 ./scripts/check-release.sh
```

No credential autodiscovery is performed by the SDK; reading from `~/.pi/agent/auth.json` is a manual test-run step only.

Note: fake-harness e2e tests remain for deterministic fault injection
(process death timing, late async failure ordering, backpressure races).

## Compatibility

- Tested with `pi-coding-agent 0.54.2`

## License

AGPL-3.0
