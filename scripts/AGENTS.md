# scripts AGENTS

## What happens here (plain words)
Repo quality gate entrypoints.

## User-facing effect
Defines what “safe to merge” and “safe to release” means for this repo.

## Happy-path flow
1. `check.sh` runs fix/fmt/vet/tests/race,
2. runs smoke example against local `pi`,
3. `check-release.sh` adds mandatory real-pi integration tests.

## Files + ownership
- `check.sh` — full local gate.
- `check-release.sh` — release gate (`PI_REAL=1` + real integration tests).
- `smoke.sh` — runtime smoke check.

## Failure/invariant notes
- `check.sh` must stay deterministic and fail-fast.
- `check-release.sh` must fail if real-pi prerequisites are missing.
- smoke should validate real executable path (`pi` in PATH).

## Canonical commands
- `./scripts/check.sh`
- `PI_REAL=1 ./scripts/check-release.sh`
- local real-cred release gate from `~/.pi/agent/auth.json` (manual explicit export):
  - `ANTHROPIC_OAUTH_TOKEN="$(python -c 'import json, pathlib; p=pathlib.Path("~/.pi/agent/auth.json").expanduser(); print(json.loads(p.read_text()).get("anthropic", {}).get("access", ""), end="")')" PI_REAL=1 ./scripts/check-release.sh`
- `./scripts/smoke.sh`
