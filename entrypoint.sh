#!/bin/sh

# 合并 GitHub 数据到 Volume（只复制不存在的文件，不覆盖已有的）

echo "📦 同步 GitHub 数据到 Volume..."

# 如果有备份的 GitHub 数据，合并到 Volume
if [ -d "/app/github-content" ]; then
    # 复制不存在的文件（-n 不覆盖）
    cp -rn /app/github-content/* /app/content/ 2>/dev/null || true
    echo "✅ content 同步完成"
fi

if [ -d "/app/github-data" ]; then
    cp -rn /app/github-data/* /app/data/ 2>/dev/null || true
    echo "✅ data 同步完成"
fi

if [ -d "/app/github-uploads" ]; then
    cp -rn /app/github-uploads/* /app/uploads/ 2>/dev/null || true
    echo "✅ uploads 同步完成"
fi

# 确保目录存在
mkdir -p /app/content/blog/default /app/content/page /app/data /app/uploads

echo "🚀 启动 mdblog..."
exec ./mdblog
