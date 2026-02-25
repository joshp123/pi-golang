# internal/sdk/tests AGENTS

## What happens here (plain words)
Black-box tests for user-visible SDK behavior.

These tests treat SDK as an external caller would: start client, send prompt,
read events, assert contracts.

## User-facing effect
Protects public guarantees:
- run lifecycle,
- abort/cancel semantics,
- process-died/close behavior,
- terminal outcome decoding.

## Happy-path flow
1. choose test mode:
   - fake harness for deterministic fault injection,
   - real pi integration for protocol truth.
2. call public SDK API,
3. consume output/events,
4. assert behavior contract.

## Files + ownership
- `e2e_helper_test.go` — helper-process bootstrap.
- `e2e_test.go` — deterministic fault-injection invariants (fake pi harness).
- `terminal_outcome_test.go` — canonical terminal status mapping.
- `real_pi_integration_test.go` — real pi integration smoke (`-tags=integration`, `PI_REAL=1`; strict mode via `PI_REAL_REQUIRED=1`). Uses explicit credentials from env or `<KEY>_FILE` env vars (e.g. `ANTHROPIC_API_KEY_FILE=/run/agenix/anthropic-api-key`); no credential autodiscovery.

## Canonical commands
- `go test ./internal/sdk/tests`
- `PI_REAL=1 go test -tags=integration ./internal/sdk/tests -run TestRealPI -v`
- `PI_REAL=1 ./scripts/check-release.sh`  # mandatory real-pi release gate
- local real-cred gate from `~/.pi/agent/auth.json` (explicit export):
  - `ANTHROPIC_OAUTH_TOKEN="$(python -c 'import json, pathlib; p=pathlib.Path("~/.pi/agent/auth.json").expanduser(); print(json.loads(p.read_text()).get("anthropic", {}).get("access", ""), end="")')" PI_REAL=1 ./scripts/check-release.sh`
- `go test ./...`
