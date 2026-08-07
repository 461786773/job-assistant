#!/usr/bin/env bash
# 打包腾讯云可部署产物（Linux amd64）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="${HOME}/.local/go/bin:${PATH}"
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"

VERSION="${VERSION:-$(date +%Y%m%d%H%M)}"
OUT_DIR="${OUT_DIR:-$ROOT/dist-release}"
PKG_NAME="job-assistant-linux-amd64-${VERSION}"
STAGE="$OUT_DIR/$PKG_NAME"
ARCH="${GOARCH:-amd64}"
OS="${GOOS:-linux}"

echo "==> clean $STAGE"
rm -rf "$STAGE"
mkdir -p "$STAGE/bin" "$STAGE/web" "$STAGE/data/uploads"

echo "==> build frontend"
cd "$ROOT/apps/web"
npm install --no-fund --no-audit
npm run build
cp -R "$ROOT/apps/web/dist/." "$STAGE/web/"

echo "==> build backend ($OS/$ARCH)"
cd "$ROOT/apps/api"
CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build -ldflags="-s -w" -o "$STAGE/bin/job-assistant" ./cmd/server

echo "==> copy runtime files"
cp "$ROOT/scripts/deploy/env.example" "$STAGE/.env.example"
cp "$ROOT/scripts/deploy/start.sh" "$STAGE/start.sh"
cp "$ROOT/scripts/deploy/stop.sh" "$STAGE/stop.sh"
cp "$ROOT/scripts/deploy/restart.sh" "$STAGE/restart.sh"
cp "$ROOT/scripts/deploy/job-assistant.service" "$STAGE/job-assistant.service"
cp "$ROOT/docs/腾讯云部署.md" "$STAGE/README.md"
chmod +x "$STAGE/start.sh" "$STAGE/stop.sh" "$STAGE/restart.sh" "$STAGE/bin/job-assistant"

# 空数据目录占位
touch "$STAGE/data/.gitkeep"

echo "==> archive"
mkdir -p "$OUT_DIR"
TAR="$OUT_DIR/${PKG_NAME}.tar.gz"
tar -C "$OUT_DIR" -czf "$TAR" "$PKG_NAME"

echo
echo "打包完成:"
echo "  目录: $STAGE"
echo "  压缩包: $TAR"
ls -lh "$TAR"
echo
echo "上传到服务器后:"
echo "  tar -xzf ${PKG_NAME}.tar.gz && cd ${PKG_NAME}"
echo "  cp .env.example .env && vi .env"
echo "  ./start.sh"
