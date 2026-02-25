#!/usr/bin/env bash
set -euo pipefail

go fix ./...

gofmt -w .

go vet ./...

if command -v staticcheck >/dev/null 2>&1; then
	staticcheck ./...
fi

go test ./...

go test -race ./...

./scripts/smoke.sh
