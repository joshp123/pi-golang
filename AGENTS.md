# pi-golang agent notes

Hi Josh. Let’s build the clean Go SDK.

## Rationale

- **Why this repo:** Lawbot needs a Go‑native interface to pi without shelling
  out per request; Clawdbot already uses the JS SDK.
- **Why RPC:** full parity with pi features, avoids re‑implementing the Node
  stack, and keeps behavior synced with upstream.
- **Why better than shelling out:** managed process, structured events,
  allowlisted env, deterministic config isolation.

## Related repos

- `~/code/lawbot-hub` (Go, shells out in `libs/pi` today)
- `~/code/clawdbot` (JS SDK via `@mariozechner/pi-coding-agent`)

## Full SDLC golden path

1. **Plan**: confirm scope; update `README.md` architecture if needed.
2. **Build**: implement changes; keep API small and stable.
3. **Verify**: run `./scripts/check.sh` (fmt + tests + smoke).
4. **Document**: update `README.md` usage + compatibility note.
5. **Release**: tag + publish (once versioning is defined).

## Golden path for new pi versions

1. Update compatibility note in `README.md`.
2. Re-run `go test ./...`.
3. Run `examples/basic` against the updated `pi` binary.
4. Compare upstream `docs/rpc.md` for protocol changes.
5. Patch event/usage parsing; add/adjust tests.
6. Document changes in README.
