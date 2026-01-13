#!/usr/bin/env bash
set -euo pipefail

if ! command -v pi >/dev/null 2>&1; then
  echo "pi CLI not found. Install @mariozechner/pi-coding-agent." >&2
  exit 1
fi

go run ./examples/basic
