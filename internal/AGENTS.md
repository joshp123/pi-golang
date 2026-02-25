# internal AGENTS

## What happens here (plain words)
`internal/` is the private engine room.

Root package `pi` is public API façade.
`internal/*` holds implementation details we can change without breaking API.

## User-facing effect
If code here breaks, users see it as:
- failed/blocked RPC requests,
- missing/late stream events,
- wrong lifecycle behavior on close/process death.

## Happy-path flow
1. `internal/sdk` builds and runs `pi --mode rpc` child process.
2. `internal/rpc` provides canonical wire command/response terms.
3. `internal/transport` handles pending requests + event queue.
4. `internal/stream` fans out events to subscribers with backpressure policy.
5. `internal/runtime` provides generic queue/registry primitives.
6. `internal/testsupport` powers fake-pi e2e scenarios.

## Failure/invariant notes
- One canonical wire term per RPC concept.
- Close/process-died semantics must remain deterministic.
- Subscriber speed must not corrupt request/response handling.

## Directory ownership
- `sdk/` — core implementation package.
- `rpc/` — wire constants/envelopes.
- `transport/` — queue/request state plumbing.
- `stream/` — subscription fanout/backpressure logic.
- `runtime/` — generic primitives.
- `testsupport/` — hermetic fake process for tests.

## Canonical commands
- `go test ./...`
