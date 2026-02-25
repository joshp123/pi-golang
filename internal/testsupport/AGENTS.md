# testsupport AGENTS

## Purpose
E2E test harness primitives for fake `pi --mode rpc` process behavior.

## What exists now
- `setup.go` — inject fake `pi` shim into PATH for a scenario.
- `runner.go` — command loop + scenario dispatch.
- `scenarios.go` — scenario handlers and JSON response/event writers.

## What is missing
- Shared typed fixtures for event payload assertions.

## Canonical commands
- `go test ./...`
