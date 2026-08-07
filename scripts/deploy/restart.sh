#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
"$ROOT/stop.sh" || true
sleep 0.5
"$ROOT/start.sh"
