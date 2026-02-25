# pi-golang agent notes

Hi Josh. Let’s build the clean Go SDK.

## Rationale

- **Why this repo:** Lawbot needs a Go‑native interface to pi without shelling
  out per request; Clawdbot already uses the JS SDK.
- **Why RPC:** full parity with pi features, avoids re‑implementing the Node
  stack, and keeps behavior synced with upstream.
- **Why better than shelling out:** managed process, structured events,
  allowlisted env, deterministic config isolation.

## Architecture principles (encoded here)

- Ontology-first naming: one canonical term per concept; no synonym drift in core paths.
- User-mechanics-first: start/control/observe lifecycle drives package boundaries.
- Tree mirrors architecture: root API façade + `internal/sdk` implementation + `internal/rpc`, `internal/transport`, `internal/stream`, `internal/runtime`.
- KISS/YAGNI: keep mirror layer thin; batteries layer separate; avoid compatibility shims.
- Discoverability-first: each non-trivial internal directory has local `AGENTS.md`.

## AGENTS.md quality bar (Feynman style, mandatory)

Every AGENTS.md must let a new engineer explain the directory after one read.

Required sections:
1. **What happens here (plain words)** — no jargon-only descriptions.
2. **User-facing effect** — what behavior changes for callers/users if this code changes.
3. **Happy-path flow** — 3-7 bullets, concrete sequence.
4. **Failure/invariant notes** — what must always stay true.
5. **Files + ownership** — where to edit for each concern.

Fail conditions:
- “wrapper/primitives/runtime” language without concrete behavior.
- Terms used but not defined (e.g., event names/modes with no examples).
- Reader cannot answer “what is this for?” in <60 seconds.

## Related repos

- `~/code/lawbot-hub` (Go, shells out in `libs/pi` today)
- `~/code/clawdbot` (JS SDK via `@mariozechner/pi-coding-agent`)

## Full SDLC golden path

1. **Plan**: confirm scope; update `README.md` architecture if needed.
2. **Build**: implement changes; keep API small and stable.
3. **Verify**: run `./scripts/check.sh` (fmt + tests + smoke).
   - Release verification: `PI_REAL=1 ./scripts/check-release.sh` (mandatory real-pi integration).
4. **Document**: update `README.md` usage + compatibility note.
5. **Release**: tag + publish (once versioning is defined).

## Commit rule

- Always run `./scripts/check.sh` before committing.

## Golden path for new pi versions

1. Update compatibility note in `README.md`.
2. Re-run `go test ./...`.
3. Run `examples/basic` against the updated `pi` binary.
4. Compare upstream `docs/rpc.md` for protocol changes.
5. Patch event/usage parsing; add/adjust tests.
6. Document changes in README.
