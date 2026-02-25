# internal/stream AGENTS

## What happens here (plain words)
Queue/fanout engine behind `client.Subscribe(...)`.

A subscriber gets a Go channel of runtime events (`message_update`, `agent_end`,
`process_died`, etc). This package defines what happens when consumers are slow.

## User-facing effect
`Subscribe(policy)` mode semantics:
- `drop` — drop new events when full (low latency, lossy).
- `block` — wait for consumer room (lossless, can stall producer path).
- `ring` — drop oldest, keep newest (best for UIs showing latest state).

Optional diagnostic: emit synthetic `subscription_drop` events.

## Happy-path flow
1. caller subscribes with buffer/mode policy.
2. hub fans each event to all subscribers.
3. each subscriber queue applies its mode semantics.
4. on process death, hub publishes `process_died` once, then closes channels.

## Failure/invariant notes
- slow subscriber must not corrupt response routing.
- process-died close behavior must be deterministic.
- mode semantics must stay stable (`drop/block/ring`).

## Files + ownership
- `policy.go` — internal policy type.
- `subscription.go` — per-subscriber queue semantics.
- `hub.go` — multi-subscriber fanout + lifecycle.

## Canonical commands
- `go test ./...`
