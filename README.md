# pi-golang

Go SDK for the pi coding agent with a clean, managed RPC process.

## Architecture

pi-golang is a thin Go wrapper around pi’s **RPC mode**. It does not re-implement
pi’s Node internals. It runs a long‑lived `pi --mode rpc` process and communicates
via JSON lines on stdin/stdout.

**Components**
- **Command resolver**: locates `pi` or a `cli.js` entrypoint (Node fallback).
- **Process manager**: spawns a single child process, manages lifecycle, restarts.
- **RPC transport**: request/response routing with IDs; event stream channel.
- **Event helpers**: extract assistant text + usage from `agent_end` events.
- **Env isolation**: allowlisted env + isolated `PI_CODING_AGENT_DIR`.

## Why RPC (and not Go reimplementation)
- pi’s “batteries” live in Node: `pi-ai`, `pi-agent-core`, `pi-coding-agent`.
- A native Go SDK would require re‑implementing tools, sessions, skills, models,
  compaction, usage, and tool loops.
- RPC gives **full parity** with pi features and keeps behavior in sync.

## Related projects
- **Lawbot**: currently shells out to pi from Go (`lawbot-hub/libs/pi`).
- **Clawdbot**: uses the in‑process JS SDK (`@mariozechner/pi-coding-agent`).
- **Goal**: pi-golang provides a **clean Go surface** with isolated env +
  long‑lived process management — no per‑call shelling and no env pollution.

## Quick start

```go
client, err := pi.Start(pi.Options{})
if err != nil {
    // handle
}
defer client.Close()

result, err := client.Run(context.Background(), "Summarize this.")
if err != nil {
    // handle
}
fmt.Println(result.Text)
```

## Pre-commit checks

Run:

```bash
./scripts/check.sh
```

Includes `gofmt`, `go test ./...`, and the smoke test. The smoke test runs
`go run ./examples/basic` and requires the `pi` CLI plus auth configured.

## Upgrade golden path

Pi versions ship frequent breaking changes. Keep a clear, repeatable upgrade:

1. **Pin + record**: update the compatibility note below.
2. **Run contract tests**: `go test ./...`.
3. **Validate RPC**: run `examples/basic` against the new `pi` binary.
4. **Diff protocol**: compare `docs/rpc.md` in `pi-mono` for changed fields.
5. **Update parsing**: adjust event/usage parsing in `pi` package.
6. **Document**: update README compatibility + changelog.

### Compatibility

- **Tested with**: _TBD_ (fill after first working release)

## License

AGPL-3.0
