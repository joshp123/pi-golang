#!/usr/bin/env bash
set -euo pipefail

if [[ "${PI_REAL:-}" != "1" ]]; then
	echo "release gate requires PI_REAL=1" >&2
	exit 1
fi

./scripts/check.sh

PI_REAL_REQUIRED=1 go test -tags=integration ./internal/sdk/tests -run TestRealPI -count=1 -v
