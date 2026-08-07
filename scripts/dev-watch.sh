#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="${HOME}/.local/go/bin:${PATH}"
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"

export JA_ROOT="$ROOT"
export JA_DATA_DIR="${JA_DATA_DIR:-$ROOT/data}"
export JA_UPLOAD_DIR="${JA_UPLOAD_DIR:-$ROOT/data/uploads}"
export JA_ADDR="${JA_ADDR:-:8080}"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

mkdir -p "$JA_DATA_DIR" "$JA_UPLOAD_DIR" "$ROOT/bin"

cd "$ROOT/apps/api"
go mod tidy
go build -o "$ROOT/bin/job-assistant" ./cmd/server

"$ROOT/bin/job-assistant" &
API_PID=$!
trap 'kill $API_PID 2>/dev/null || true' EXIT

cd "$ROOT/apps/web"
npm install
echo "API  http://127.0.0.1${JA_ADDR}"
echo "Web  http://127.0.0.1:5173  (proxy /api -> API)"
npm run dev -- --host 0.0.0.0 --port 5173
