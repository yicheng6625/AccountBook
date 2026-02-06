#!/bin/bash
# 建置腳本：本機編譯後端二進位檔
# 用法：在專案根目錄執行 ./build.sh，然後再 docker compose up -d --build

set -e

echo "=== 開始編譯後端 ==="

PROJECT_DIR="/Users/steven/Documents/GitHub/1_AccountBook"

cd "$PROJECT_DIR"

git pull origin main

cd backend

# 安裝 Go 依賴
go mod tidy

# 交叉編譯為 Linux/ARM64（Docker 容器環境）
# 原因：使用純 Go SQLite 驅動，不需要 CGO
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=arm64 \
go build -o accountbook-server .

echo "=== 後端編譯完成 ==="
echo ""
echo "接下來將執行："
echo "  docker compose up -d --build"

docker compose up -d --build