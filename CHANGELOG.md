# Changelog

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
