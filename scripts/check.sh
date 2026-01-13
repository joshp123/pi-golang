#!/usr/bin/env bash
set -euo pipefail

gofmt -w .

go test ./...

./scripts/smoke.sh
