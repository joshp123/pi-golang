# internal/testsupport AGENTS

## What happens here (plain words)
Fake `pi` process used by e2e tests.

Tests put a shim `pi` executable in PATH.
That shim runs test helper code from this directory and emits scripted RPC
responses/events.

## User-facing effect
Gives confidence that public SDK contracts hold under real process I/O patterns:
- async prompt failures,
- process death,
- run cancellation/abort,
- event floods/backpressure.

## Happy-path flow
1. test picks scenario name (`happy`, `die_on_prompt`, ...).
2. `setup.go` injects fake `pi` shim into PATH.
3. helper process enters read-command/write-response loop.
4. `scenarios.go` emits deterministic wire frames for that scenario.
5. SDK e2e test asserts public behavior.

## Failure/invariant notes
- scenario outputs must stay deterministic.
- wire envelopes must mirror real RPC shape.
- helper must fail loudly on unknown scenario.

## Files + ownership
- `setup.go` — fake `pi` shim setup.
- `runner.go` — stdin command loop + scenario dispatch.
- `scenarios.go` — scripted behavior per scenario.

## Canonical commands
- `go test ./...`
