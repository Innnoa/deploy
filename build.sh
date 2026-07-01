#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_FILE="$SCRIPT_DIR/build-$(date +%Y%m%d-%H%M%S).log"
exec > >(tee -a "$LOG_FILE") 2>&1

echo "=========================================="
echo "Deploy Build Script"
echo "Started: $(date '+%Y-%m-%d %H:%M:%S')"
echo "Log: $LOG_FILE"
echo "=========================================="

# 1. Check prerequisites
echo ""
echo "[1/4] Checking environment..."
export PATH=$PATH:/usr/local/go/bin:/root/go/bin

command -v go >/dev/null || { echo "ERROR: go not found"; exit 1; }
command -v wails >/dev/null || { echo "ERROR: wails not found"; exit 1; }

echo "  Go:    $(go version)"
echo "  Wails: $(wails version 2>/dev/null | head -1)"
echo "  Node:  $(node -v 2>/dev/null || echo 'not found')"

# 2. Set Wails version
echo ""
echo "[2/4] Setting Wails v2.4.0 for Kylin compatibility..."
sed -i 's|github.com/wailsapp/wails/v2 v2.10.1|github.com/wailsapp/wails/v2 v2.4.0|' go.mod
grep 'wailsapp/wails/v2 v' go.mod

# 3. Tidy dependencies
echo ""
echo "[3/4] Tidying Go modules..."
go mod tidy
echo "  Done."

# 4. Build
echo ""
echo "[4/4] Building for linux/amd64..."
START_TIME=$(date +%s)

wails build \
  -platform linux/amd64 \
  -clean \
  -ldflags "-s -w -X main.Version=1.0.14 -X main.BaseUrl=https://deploy.ru.com/api-system" \
  -o Deploy-kylin

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""
echo "=========================================="
echo "Build completed: $(date '+%Y-%m-%d %H:%M:%S')"
echo "Duration: ${DURATION}s"
echo "Binary: $(ls -lh build/bin/Deploy-kylin | awk '{print $5, $NF}')"
echo "Log: $LOG_FILE"
echo "=========================================="
