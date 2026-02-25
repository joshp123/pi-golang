# internal/sdk AGENTS

## What happens here (plain words)
This is the actual SDK engine behind public package `pi`.

Root `pi` files mostly re-export. Real behavior lives here.

## User-facing effect
Changes here directly change:
- request/response behavior,
- run/abort semantics,
- event streaming and close/process-died guarantees,
- typed decoder behavior.

## Happy-path flow
1. `StartSession` / `StartOneShot` starts `pi --mode rpc` child process.
2. thin mirror methods build canonical RPC command JSON.
3. `send.go` writes command + waits for matching response ID.
4. `process.go` parses stdout lines into responses/events.
5. `subscriptions.go` routes events to subscribers per policy.
6. batteries layer (`Run*`, `ShareSession`) composes mirror calls.

## Failure/invariant notes
- mirror methods stay 1:1 with upstream RPC contracts.
- deterministic behavior on close/process-died.
- single-flight `Run` invariant (`ErrRunInProgress`).
- strict decoder envelope checks (no silent fallback).
- child-process credentials/environment are explicit by default (`InheritEnvironment=false`).

## Layout
- `client.go` — lifecycle + process management.
- `api_rpc.go` / `rpc_contracts.go` — thin mirror + command/response contracts.
- `send.go` / `process.go` / `subscriptions.go` — transport/event runtime.
- `batteries_*.go` — convenience APIs.
- `decode.go` — typed event/terminal decoders.
- `types.go` / `errors.go` / `mode.go` / `options.go` / `env.go` / `command.go` — ontology/config.

## Tests
- `internal/sdk/tests/` — black-box user behavior.
- `internal/sdk/*_test.go` — white-box runtime invariants.

## Canonical commands
- `go test ./...`
- `./scripts/check.sh`
