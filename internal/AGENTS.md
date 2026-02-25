# internal AGENTS

## Purpose
Private implementation packages for test/runtime support.

## What exists now
- `runtime/` — generic runtime primitives used by root package internals.
- `testsupport/` — hermetic fake-pi helper runtime used by e2e tests.

## What is missing
- Further extraction only when root package clarity improves from it.

## Canonical commands
- `go test ./...`
