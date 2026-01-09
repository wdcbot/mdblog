# mdblog

一个基于 Go 语言开发的、高性能、无数据库的轻量级 Markdown 博客系统。

> 本项目基于 [TwoThreeWang/mdblog](https://github.com/TwoThreeWang/mdblog) 二次开发，进行了大量功能增强和优化。

## ✨ 特性

- **无数据库** - 文章以 Markdown 文件存储，评论用 JSON 文件
- **高性能** - 内存缓存 + Gzip 压缩 + 静态资源缓存
- **全文搜索** - 内置 Bleve 搜索引擎
- **SEO 友好** - 自动生成 Sitemap、RSS、Open Graph 标签
- **暗色模式** - 支持手动切换亮/暗主题
- **文章目录** - 长文章自动生成侧边 TOC
- **评论系统** - 内置简单评论，支持审核
- **草稿功能** - 支持保存草稿，前台不显示
- **管理后台** - 在线编辑文章、管理分类/标签/评论

## 🚀 快速开始

### 1. 克隆项目
```bash
git clone https://github.com/wdcbot/mdblog.git
cd mdblog
```

### 2. 配置环境变量
```bash
cp .env.example .env
```

编辑 `.env` 文件：
```env
PORT=8080
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_password
JWT_SECRET=your_secret_key
```

### 3. 运行
```bash
go run main.go
```

### 4. 访问
- 前台：http://localhost:8080
- 后台：http://localhost:8080/admin

## � 项目结构

```
mdblog/
├── admin/              # 后台管理
│   ├── layouts/        # 后台模板
│   └── static/         # 后台静态资源
├── content/            # 内容目录
│   ├── blog/           # 博客文章（按分类存放）
│   └── page/           # 独立页面
├── data/               # 数据目录
│   └── comments.json   # 评论数据
├── internal/           # 核心代码
│   ├── pkg/            # 功能模块
│   ├── router/         # 路由
│   └── theme/          # 模板渲染
├── themes/             # 主题目录
│   └── pure/           # 默认主题
├── config.yaml         # 站点配置
└── main.go             # 入口文件
```

## 📝 文章格式

```markdown
---
title: "文章标题"
date: 2026-01-10
tags: [Go, 博客]
draft: false
---

文章内容...
```

## 🛠️ 技术栈

- **后端**: Go + Gin
- **模板**: Pongo2 (前台) + html/template (后台)
- **Markdown**: Goldmark
- **搜索**: Bleve
- **前端**: 原生 CSS + HTMX

## 📄 License

MIT License
