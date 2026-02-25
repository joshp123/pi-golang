# internal/rpc AGENTS

## What happens here (plain words)
Defines exact JSON wire terms used between SDK and upstream `pi --mode rpc`.

If a constant in `wire.go` changes, protocol behavior changes.

## User-facing effect
- Thin mirror methods (`Prompt`, `Abort`, `GetState`, ...) rely on these names.
- Wrong wire terms = broken RPC calls or decode failures.

## Happy-path flow
1. SDK builds `Command` map using canonical command constants.
2. command is marshaled and sent to pi process.
3. pi responds with `Response` envelope shape defined here.
4. runtime decodes and routes by response ID/type.

## Failure/invariant notes
- mirror upstream `docs/rpc.md` exactly.
- no alias/synonym constants in core wire set.
- keep one canonical term per concept.

## Files + ownership
- `wire.go`
  - `Command` / `Response` wire types.
  - command constants (`prompt`, `abort`, ...).
  - transport response/parse event constants.

## Canonical commands
- `go test ./...`
