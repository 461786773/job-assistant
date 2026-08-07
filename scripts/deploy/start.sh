#!/usr/bin/env bash
# 生产启动（腾讯云 CVM / 任意 Linux）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

if [[ ! -f "$ROOT/.env" ]]; then
  echo "缺少 .env，请先: cp .env.example .env 并填写 JA_LLM_*"
  exit 1
fi

set -a
# shellcheck disable=SC1091
source "$ROOT/.env"
set +a

export JA_ROOT="${JA_ROOT:-$ROOT}"
export JA_DATA_DIR="${JA_DATA_DIR:-$ROOT/data}"
export JA_UPLOAD_DIR="${JA_UPLOAD_DIR:-$ROOT/data/uploads}"
export JA_WEB_DIR="${JA_WEB_DIR:-$ROOT/web}"
export JA_ADDR="${JA_ADDR:-:8080}"

mkdir -p "$JA_DATA_DIR" "$JA_UPLOAD_DIR"

if [[ -f "$ROOT/data/app.pid" ]]; then
  OLD="$(cat "$ROOT/data/app.pid" || true)"
  if [[ -n "${OLD}" ]] && kill -0 "$OLD" 2>/dev/null; then
    echo "已在运行 pid=$OLD，如需重启请执行 ./restart.sh"
    exit 0
  fi
fi

if [[ ! -x "$ROOT/bin/job-assistant" ]]; then
  echo "找不到可执行文件: $ROOT/bin/job-assistant"
  exit 1
fi

nohup "$ROOT/bin/job-assistant" >> "$ROOT/data/app.log" 2>&1 &
echo $! > "$ROOT/data/app.pid"
sleep 0.5

if kill -0 "$(cat "$ROOT/data/app.pid")" 2>/dev/null; then
  echo "启动成功 pid=$(cat "$ROOT/data/app.pid")"
  echo "监听: http://0.0.0.0${JA_ADDR}"
  echo "日志: $ROOT/data/app.log"
  echo "健康检查: curl http://127.0.0.1${JA_ADDR}/api/health"
else
  echo "启动失败，请查看 $ROOT/data/app.log"
  exit 1
fi
