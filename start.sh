#!/bin/bash

# mdblog 快速启动脚本

set -e

echo "🚀 mdblog"
echo "========="

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ 未检测到 Go，请先安装: https://golang.org/dl/"
    echo "   或使用 Docker: docker-compose up -d"
    exit 1
fi

# 创建必要目录
mkdir -p content/blog/default content/page data uploads

# 运行
echo "✅ 启动中..."
go run main.go
