# pi-golang

Go SDK for the pi coding agent using a managed `pi --mode rpc` process.

Thanks to Mario Zechner and the pi project: [pi-mono](https://github.com/badlogic/pi-mono).

## Design goals

- One process per client (no per-call shelling)
- One obvious request shape (`PromptRequest`) for all user messages
- Two explicit layers:
  - thin RPC mirror (typed methods, 1:1 with upstream commands)
  - batteries (ergonomic helpers like `Run`, decoders, subscription policies)
- Explicit failure semantics (`ErrProcessDied`, `ErrClientClosed`, `ErrRunInProgress`)
- Strict contracts over silent fallbacks

## Quick start

```go
opts := pi.DefaultOneShotOptions()
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
- Typed event decoders (`DecodeAgentEnd`, `DecodeMessageUpdate`, `DecodeAutoCompactionEnd`)
- `ShareSession(ctx)` (export + gist helper)

## Intent map (ontology-first)

- Ask: `Prompt`, `Run`
- Steer queued work: `Steer`, `FollowUp`
- Abort current work: `Abort`
- Inspect runtime/session: `GetState`, `Stderr`
- Compact/session mgmt: `Compact`, `NewSession`, `ExportHTML`, `ShareSession`
- Observe stream: `Subscribe` + typed event decoders
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

## Event subscription

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
- `drop`
- `block`
- `ring`

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

## API contract

- Contexts:
  - RPC methods require non-nil context (`ErrNilContext`).
  - If context has no deadline, a default 2m timeout is applied.
- Thin mirror methods (`Prompt`, `Steer`, `FollowUp`, `Abort`, `GetState`, `NewSession`, `Compact`, `ExportHTML`) map 1:1 to upstream RPC commands.
- `Run` is a battery helper:
  - single-flight per client (`ErrRunInProgress` on overlap)
  - sends one `prompt`, waits for `agent_end`
  - on context cancellation while waiting, sends best-effort `Abort` and returns `ctx.Err()`
  - surfaces late async `prompt` failures (`response` frames) as `*RPCError`
- `Abort(ctx)` sends upstream `{"type":"abort"}` and waits for command response.
- Process/lifecycle guarantees:
  - unexpected process exit fails pending requests with `ErrProcessDied`
  - emits exactly one `process_died` event
  - closes all subscriber channels after that event
  - `Close()` deterministically unblocks pending requests with `ErrClientClosed`
- Decoder strictness: RPC/event payloads must include explicit `type` values matching the expected envelope; missing/mismatched types fail fast.
- Raw transport path is internal (`send` + `rpcCommand`/`rpcResponse` are not public API).

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

## Compatibility

- Tested with `pi-coding-agent 0.54.2`

## License

AGPL-3.0
