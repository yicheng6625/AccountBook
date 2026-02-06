#!/bin/bash
# AccountBook 部署腳本
# 原因：供 GitHub Actions self-hosted runner 呼叫，自動拉取最新程式碼並部署

set -e  # 遇到錯誤立即停止

# 專案路徑（請依實際情況修改）
PROJECT_DIR="/Users/steven/Documents/GitHub/1_AccountBook"

echo "=========================================="
echo "開始部署 AccountBook"
echo "時間：$(date '+%Y-%m-%d %H:%M:%S')"
echo "=========================================="

# 進入專案目錄
cd "$PROJECT_DIR"

# 拉取最新程式碼
echo ">>> 拉取最新程式碼..."
git pull origin main

# 進入後端目錄，下載 Go 依賴
echo ">>> 下載 Go 依賴..."
cd "$PROJECT_DIR/backend"
# go mod download
go mod tidy

# 編譯後端（arm64 for Mac mini M 系列，若為 Intel 請改為 amd64）
echo ">>> 編譯後端..."
# 交叉編譯為 Linux/ARM64（Docker 容器環境）
# 原因：使用純 Go SQLite 驅動，不需要 CGO
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=arm64 \
go build -o accountbook-server .

chmod +x accountbook-server

# 回到專案根目錄
# cd "$PROJECT_DIR"

# 停止舊容器、重新建置並啟動
echo ">>> 重新建置並啟動 Docker 容器..."
# docker compose down
# 先手動把程式碼包進新的鏡像，但強制使用本地現有的基礎鏡像
docker compose build --pull=false
docker compose up -d

# 清理舊映像
echo ">>> 清理未使用的 Docker 映像..."
docker image prune -f

# 顯示容器狀態
echo ">>> 目前容器狀態："
docker compose ps

echo "=========================================="
echo "部署完成！"
echo "時間：$(date '+%Y-%m-%d %H:%M:%S')"
echo "=========================================="
