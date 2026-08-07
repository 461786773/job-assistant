#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="${HOME}/.local/go/bin:${PATH}"
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
export JA_ROOT="$ROOT"
export JA_DATA_DIR="${JA_DATA_DIR:-$ROOT/data}"
export JA_UPLOAD_DIR="${JA_UPLOAD_DIR:-$ROOT/data/uploads}"
export JA_WEB_DIR="${JA_WEB_DIR:-$ROOT/apps/web/dist}"
export JA_ADDR="${JA_ADDR:-:8080}"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

mkdir -p "$JA_DATA_DIR" "$JA_UPLOAD_DIR" "$ROOT/bin"

echo "==> building web"
cd "$ROOT/apps/web"
npm install
npm run build

echo "==> building api"
cd "$ROOT/apps/api"
go mod tidy
go build -o "$ROOT/bin/job-assistant" ./cmd/server

echo "==> starting on ${JA_ADDR}"
exec "$ROOT/bin/job-assistant"
