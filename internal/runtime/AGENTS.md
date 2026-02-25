# internal/runtime AGENTS

## What happens here (plain words)
Tiny generic concurrency building blocks used by higher internal packages.

Think:
- one safe pending-request registry,
- one blocking queue with close semantics.

## User-facing effect
Indirect but critical:
- pending RPC calls resolve/close correctly,
- event dispatch wakes/sleeps/terminates predictably.

## Happy-path flow
1. `PendingRegistry` stores requestID -> response channel.
2. response arrives, registry resolves channel and closes it.
3. process dies/close path marks terminal error and closes all pending channels.
4. `Queue` buffers events until consumer pops them.
5. queue close unblocks waiters and returns `ok=false`.

## Failure/invariant notes
- no leaked channels on close/process death.
- no post-close enqueue success.
- terminal error, once set, stays authoritative.

## Files + ownership
- `pending.go` — pending request registry state machine.
- `queue.go` — blocking FIFO with close.

## Canonical commands
- `go test ./...`
