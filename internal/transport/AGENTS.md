# internal/transport AGENTS

## What happens here (plain words)
Low-level request/event plumbing used by `internal/sdk` runtime loops.

This package handles synchronization/state transitions, not product-level logic.

## User-facing effect
- RPC requests resolve correctly (or fail deterministically on close/process death).
- Event queue prevents subscriber speed from breaking response parsing flow.

## Happy-path flow
1. request is registered in pending registry by request ID.
2. stdout parser decodes matching response.
3. registry resolves channel and closes it.
4. non-response events go through queue to subscriber fanout path.

## Failure/invariant notes
- terminal process error must close all pending request channels.
- post-close request registration must fail.
- queue close must unblock waiters.

## Files + ownership
- `requests.go`
  - pending registry wrapper + terminal error tracking.
- `queue.go`
  - typed blocking FIFO wrapper.

## Canonical commands
- `go test ./...`
