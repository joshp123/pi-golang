# examples AGENTS

## What happens here (plain words)
Runnable programs showing how to call the public SDK.

## User-facing effect
Examples are copy/paste onboarding paths.
If examples drift from API reality, users integrate wrong.

## Happy-path flow
1. create options,
2. start client,
3. run prompt,
4. print result,
5. close client.

## Files + ownership
- `basic/main.go` — smallest one-shot flow.

## Missing next
- session example (`StartSession`, `ExportHTML`, `ShareSession`).
- stream example (`Subscribe` with `drop/block/ring`).

## Canonical commands
- `go run ./examples/basic`
- `go test ./...`
