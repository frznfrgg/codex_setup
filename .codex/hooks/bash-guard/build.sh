#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
SRC="$ROOT/src"
REQUIRED_GO="$(awk '/^go / { print $2; exit }' "$SRC/go.mod")"

if ! command -v go >/dev/null 2>&1; then
  echo "go is required to build bash-guard. Install Go ${REQUIRED_GO}+ and rerun this script." >&2
  exit 1
fi

cd "$SRC"
go test -count=1 ./...
go build -o bash_guard.bin .
./bash_guard.bin --selftest
