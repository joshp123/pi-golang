# runtime AGENTS

## Purpose
Private runtime primitives extracted from root package for clearer architecture.

## What exists now
- `pending.go` — generic pending-request registry with terminal state ownership.
- `queue.go` — generic blocking FIFO with close semantics.

## What is missing
- Additional runtime primitives if fanout/process lifecycle needs deeper extraction.

## Canonical commands
- `go test ./...`
