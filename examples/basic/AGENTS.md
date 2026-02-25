# examples/basic AGENTS

## What happens here (plain words)
Minimal program that starts one-shot client, sends one prompt, prints assistant
text.

## User-facing effect
Fast sanity check that local `pi` CLI + SDK wiring works end-to-end.

## Happy-path flow
1. build default options,
2. `StartOneShot`,
3. `Run(context.Background(), PromptRequest{...})`,
4. print text,
5. close client.

## Files + ownership
- `main.go`

## Canonical commands
- `go run ./examples/basic`
