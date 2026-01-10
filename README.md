# mdblog

轻量级 Markdown 博客系统，无数据库依赖。

> 本项目基于 [TwoThreeWang/mdblog](https://github.com/TwoThreeWang/mdblog) 二次开发，进行了大量功能增强和优化。

## 特性

- 📄 Markdown 文件存储
- ⚡ 内存缓存，高性能
- 🔍 内置全文搜索
- 🌙 暗色模式
- 💬 评论系统
- 📱 响应式设计
- 🎨 后台可视化编辑

## 快速部署到 Zeabur

1. Fork 这个仓库到你的 GitHub
2. 去 [Zeabur](https://zeabur.com) 用 GitHub 登录
3. 创建项目 → 选择「共享集群」→ 选择「香港」地区
4. 添加服务 → Git → 选择你 fork 的仓库
5. 等待构建完成，自动获得访问域名

每次 push 代码，Zeabur 会自动重新部署。

## 本地运行

```bash
git clone https://github.com/wdcbot/mdblog.git
cd mdblog
go run main.go
```

- 前台：http://localhost:8080
- 后台：http://localhost:8080/admin（默认 admin / admin888）

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

## 配置

编辑 `config.yaml`：

```yaml
server:
    port: 8080

admin:
    username: admin
    password: 你的密码

site:
    title: 我的博客
    description: 博客描述
```

## 目录结构

```
mdblog/
├── content/blog/    # 博客文章
├── content/page/    # 独立页面
├── data/            # 评论、统计
├── uploads/         # 上传图片
├── config.yaml      # 配置文件
└── main.go
```

## License

MIT
