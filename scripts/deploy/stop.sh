#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
PID_FILE="$ROOT/data/app.pid"

if [[ ! -f "$PID_FILE" ]]; then
  echo "未找到 pid 文件，服务可能未启动"
  exit 0
fi

PID="$(cat "$PID_FILE" || true)"
if [[ -z "$PID" ]]; then
  rm -f "$PID_FILE"
  echo "pid 为空，已清理"
  exit 0
fi

if kill -0 "$PID" 2>/dev/null; then
  kill "$PID" 2>/dev/null || true
  sleep 0.5
  if kill -0 "$PID" 2>/dev/null; then
    kill -9 "$PID" 2>/dev/null || true
  fi
  echo "已停止 pid=$PID"
else
  echo "进程不存在 pid=$PID"
fi
rm -f "$PID_FILE"
