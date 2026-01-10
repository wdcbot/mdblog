#!/bin/sh

# 合并 GitHub 数据到 Volume（只复制不存在的文件，不覆盖已有的）

echo "📦 同步 GitHub 数据到 Volume..."

# 确保目录存在
mkdir -p /app/content/blog /app/content/page /app/data /app/uploads

# 同步 content 目录
if [ -d "/app/github-content" ]; then
    echo "检查 content 目录..."
    cd /app/github-content
    for file in $(find . -type f -name "*.md"); do
        target="/app/content/$file"
        if [ ! -f "$target" ]; then
            mkdir -p "$(dirname "$target")"
            cp "$file" "$target"
            echo "  新增: $file"
        else
            echo "  跳过: $file (已存在)"
        fi
    done
    cd /app
    echo "✅ content 同步完成"
fi

# 同步 data 目录
if [ -d "/app/github-data" ]; then
    echo "检查 data 目录..."
    cd /app/github-data
    for file in $(find . -type f); do
        target="/app/data/$file"
        if [ ! -f "$target" ]; then
            mkdir -p "$(dirname "$target")"
            cp "$file" "$target"
            echo "  新增: $file"
        else
            echo "  跳过: $file (已存在)"
        fi
    done
    cd /app
    echo "✅ data 同步完成"
fi

# 同步 uploads 目录
if [ -d "/app/github-uploads" ]; then
    echo "检查 uploads 目录..."
    cd /app/github-uploads
    for file in $(find . -type f ! -name ".gitkeep"); do
        target="/app/uploads/$file"
        if [ ! -f "$target" ]; then
            mkdir -p "$(dirname "$target")"
            cp "$file" "$target"
            echo "  新增: $file"
        else
            echo "  跳过: $file (已存在)"
        fi
    done
    cd /app
    echo "✅ uploads 同步完成"
fi

echo "🚀 启动 mdblog..."
exec ./mdblog
