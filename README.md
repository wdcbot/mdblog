# mdblog

轻量级 Markdown 博客系统，支持动态服务器和静态部署。

## 特性

- 📄 Markdown 文件存储
- ⚡ 支持 GitHub Pages 静态部署
- 🌙 暗色模式
- 📱 响应式设计

## 部署到 GitHub Pages

1. Fork 这个仓库
2. 修改 `config.yaml` 中的 `base_url` 为你的 GitHub Pages 地址
3. 去仓库 Settings → Pages → Source 选择 `GitHub Actions`
4. 推送代码，自动部署

每次 push 到 main 分支，GitHub Actions 会自动生成静态站点并部署。
## 本地运行

```bash
git clone https://github.com/wdcbot/mdblog.git
cd mdblog
go run main.go
```

访问 http://localhost:8080

## 写文章

在 `content/blog/分类名/` 下创建 `.md` 文件：

```markdown
---
title: "文章标题"
date: 2026-01-10
tags: [标签1, 标签2]
---

正文内容...
```

## 生成静态站点

```bash
go run main.go --build
```

生成的文件在 `public/` 目录。

## License

MIT
