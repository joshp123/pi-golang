# Changelog

## Unreleased

- Add optional `TerminalReason` passthrough on `TerminalOutcome` for upstream typed terminal reason fields when present
- Preserve strict non-heuristic behavior: no regex inference; empty `TerminalReason` when upstream omits reason details

## v0.0.14

- Add pure managed-run classifiers in batteries layer:
  - `ClassifyManaged(RunDetailedResult) -> ManagedSummary`
  - `ClassifyRunError(error) -> (BrokenCause, bool)`
- Add stable managed completion classes (`ok`, `ok_after_recovery`, `aborted`, `failed`)
- Add typed recovery facts (`CompactionObserved`, `OverflowDetected`, `Recovered`) without regex heuristics
- Keep orchestration unchanged (`Run`/`RunDetailed` remain the execution path)

## v0.0.13

- Add generic `CompactionPrompt` option to `SessionOptions` and `OneShotOptions`
- Add SDK-managed compaction hook bootstrap:
  - creates temp extension bundle + prompt file
  - starts pi with `--extension <bundle>/compaction-hook.ts`
  - injects `PI_GOLANG_COMPACTION_PROMPT_FILE` and `PI_GOLANG_COMPACTION_PROMPT_SHA256`
  - cleans up temp bundle on `Close()`
- Add tests for hook bundle generation, env/arg injection, and cleanup
- Add real-pi integration coverage to verify hook execution via compaction details marker (`source: pi-golang-compaction-prompt`)
- Document `CompactionPrompt` usage and contract in README

## v0.0.12

- Refactor package layout to root API façade + `internal/sdk` implementation package
- Add `internal/sdk/tests/` for black-box/e2e behavior tests; keep white-box invariants in `internal/sdk/*_test.go`
- Keep public API/contracts stable via root re-exports (`Client`, typed entrypoints, errors, decoders, modes/options)
- Add explicit auth/env controls on start options (`Auth`, `Environment`, `InheritEnvironment`, `SeedAuthFromHome`)
- Flip default to explicit environment (`InheritEnvironment=false`) and keep home-auth seeding configurable (`SeedAuthFromHome`)
- Remove legacy OAuth file fallback search paths; seed only from `~/.pi/agent` when enabled
- Add real pi integration tests (`internal/sdk/tests/real_pi_integration_test.go`, `-tags=integration`, `PI_REAL=1`) plus strict prereq mode (`PI_REAL_REQUIRED=1`)
- Add release gate script `scripts/check-release.sh` that requires real-pi integration pass
- Trim fake-harness e2e scope to deterministic fault-injection invariants (no broad happy-path duplication)
- Rewrite internal AGENTS docs to Feynman-style plain-language behavior explanations

## v0.0.11

- Architecture pass: move runtime concerns into explicit internal folders (`internal/rpc`, `internal/transport`, `internal/stream`)
- Remove root-level slop wrappers by centralizing event fanout/backpressure into a generic internal stream hub
- Keep thin mirror + batteries contracts stable while replacing private wire names with `internal/rpc` canonical constants
- Add local `AGENTS.md` docs for new internal architecture directories for tree-level discoverability

## v0.0.10

- Extract generic runtime primitives to `internal/runtime/` (`PendingRegistry`, `Queue`) and keep root wrappers thin
- Add `RunDetailed` battery helper to return terminal outcome plus typed compaction/retry signals
- Add typed decoders for `auto_compaction_start`, `auto_retry_start`, and `auto_retry_end`
- Preserve strict overflow stance: no copied provider-regex classifier in core until upstream RPC provides typed reason fields

## v0.0.9

- Add canonical `DecodeTerminalOutcome` for `agent_end` payloads (`completed|failed|aborted` + text/stopReason/error/usage)
- Tighten `GetState` contract: `SessionState.ContextWindow` must be present (>0) after model fallback
- Move fake-pi e2e harness into `internal/testsupport/` (setup/runner/scenarios) for clearer tree-level architecture
- Add local `AGENTS.md` docs for new internal directories
- Keep overflow handling strict: no copied provider-regex classifier in core SDK until upstream RPC exposes a typed reason

## v0.0.8

- Add typed `Abort(ctx)` with upstream RPC parity (`{"type":"abort"}`)
- Split architecture explicitly into thin RPC mirror vs batteries layer (`Run`, decoders, sharing helpers)
- Simplify pending-request lifecycle manager from actor-style supervisor to mutex-backed state machine
- Keep strict close/process-death contracts while making control flow easier to explain and audit
- Expand tests with abort unit coverage + e2e run-interrupt behavior
- Make Run cancellation semantics explicit: on ctx cancel while waiting, send best-effort abort then return ctx error
- Enforce strict envelope type decoding for RPC responses and typed event decoders (no missing-type fallback)
- Centralize canonical command/event terms in one ontology constant set and remove string drift in core paths
- Split oversized e2e test fixture into focused files to keep test files under 400 LOC and improve explainability
- Remove unused client state (`command`, `waitErr`) and keep lifecycle flow minimal (wait -> mark process died)
- Add local directory AGENTS docs for `examples/` and `scripts/` to improve tree-level discoverability

## v0.0.7

- Breaking API cleanup: remove public raw-RPC escape hatches and standardize on typed methods only
- Enforce one request shape (`PromptRequest`) across `Prompt`, `Run`, `Steer`, and `FollowUp`
- Collapse subscriptions to one API (`Subscribe(SubscriptionPolicy)`) with explicit `Buffer` + mode
- Introduce actor-style request supervisor for pending-request/process-error lifecycle ownership
- Split event fanout into a dedicated hub with deterministic shutdown + unblock behavior
- Keep async prompt-failure detection in `Run` while preserving response parsing under subscriber backpressure

## v0.0.6

- Enforce strict API contracts: nil contexts now fail with `ErrNilContext`, close paths return `ErrClientClosed`, and RPC failures return typed `*RPCError`
- Decouple stdout parsing from subscriber delivery with an internal event queue; block-mode subscribers no longer stall response parsing
- Harden `Run` semantics with single-flight protection (`ErrRunInProgress`) and detection of async prompt failures emitted as late `response` frames
- Align prompt image payloads with upstream RPC (`{type:"image", data, mimeType}`) and add `PromptWithBehavior`, `Steer`, `FollowUp`
- Make subscription policy validation explicit (`ErrInvalidSubscriptionPolicy`) instead of silent mode/buffer coercion
- Expand tests for concurrency and lifecycle invariants (concurrent run rejection, close-unblocks-pending, async prompt errors, blocked-subscriber response progress)

## v0.0.5

- Add typed runtime wrappers: `GetState`, `NewSession`, `Compact`
- Add typed event decoders: `DecodeAgentEnd`, `DecodeMessageUpdate`, `DecodeAutoCompactionEnd`
- Add configurable subscription backpressure via `SubscribeWithPolicy`
- Fail pending requests/subscribers immediately with `ErrProcessDied` on unexpected process death
- Expand hermetic e2e + policy tests for transport semantics

## v0.0.4

- Add `ShareSession` to export and gist session HTML

## v0.0.3

- Split clients into `SessionClient` and `OneShotClient` with explicit sessions
- Add `SessionOptions`/`OneShotOptions` and `SessionName` support

## v0.0.2

- Add `SessionClient`/`OneShotClient` with explicit session behavior
- Keep opinionated modes + dragons escape hatch

## v0.0.1

- RPC-based Go client with managed `pi --mode rpc` process
- Env isolation + agent dir seeding for clean auth handling
- Command resolver for `pi` / `cli.js` entrypoints
- Run helpers to extract assistant text + usage
- Example + smoke/check scripts (`./scripts/check.sh`)
