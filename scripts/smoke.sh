#!/usr/bin/env bash
set -euo pipefail

if ! command -v pi >/dev/null 2>&1; then
  echo "pi CLI not found. Install @mariozechner/pi-coding-agent." >&2
  exit 1
fi

has_creds=0
for key in ANTHROPIC_API_KEY OPENAI_API_KEY; do
  if [[ -n "${!key:-}" ]]; then
    has_creds=1
    break
  fi
  file_var="${key}_FILE"
  file_path="${!file_var:-}"
  if [[ -n "$file_path" && -s "$file_path" ]]; then
    has_creds=1
    break
  fi
done

if [[ "$has_creds" -ne 1 ]]; then
  echo "smoke: skipping (set ANTHROPIC_API_KEY[/_FILE] or OPENAI_API_KEY[/_FILE])"
  exit 0
fi

go run ./examples/basic
