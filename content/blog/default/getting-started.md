---
title: "mdblog 使用指南"
date: 2026-01-10
tags: [教程, 入门]
pinned: true
---

欢迎使用 mdblog！这是一个轻量级的 Markdown 博客系统，无需数据库，简单易用。

## 部署到 Zeabur（推荐）

最简单的部署方式，5 分钟上线：

### 1. Fork 仓库

去 [GitHub](https://github.com/wdcbot/mdblog) Fork 这个项目到你的账号。

### 2. 登录 Zeabur

访问 [zeabur.com](https://zeabur.com)，用 GitHub 账号登录。

### 3. 创建项目

1. 点击「创建新项目」
2. 选择「共享集群」（有免费试用）
3. 地区选择「香港」（国内访问快）

### 4. 部署服务

1. 点击「添加服务」→「Git」
2. 选择你 fork 的 mdblog 仓库
3. Zeabur 会自动检测 Go 项目并开始构建
4. 等待几分钟，构建完成后点击「生成域名」

### 5. 完成

访问生成的域名即可看到你的博客！

后台地址：`你的域名/admin`，默认账号 `admin`，密码 `admin888`

> 每次你 push 代码到 GitHub，Zeabur 会自动重新部署。

---

## 本地运行

如果你想在本地开发调试：

```bash
git clone https://github.com/你的用户名/mdblog.git
cd mdblog
go run main.go
```

访问：
- 前台：http://localhost:8080
- 后台：http://localhost:8080/admin

---

## 写文章

### 方式一：后台编辑器

1. 登录后台 `/admin`
2. 点击「撰写新文章」
3. 使用 Markdown 编辑器写作
4. 支持图片上传、粘贴图片
5. 点击「保存」发布

### 方式二：直接创建文件

在 `content/blog/分类名/` 下创建 `.md` 文件：

```markdown
---
title: "文章标题"
date: 2026-01-10
tags: [标签1, 标签2]
---

正文内容，支持 Markdown 语法...
```

---

## 文章属性

在文章开头的 `---` 之间设置：

| 属性 | 说明 | 示例 |
|------|------|------|
| `title` | 文章标题 | `"我的文章"` |
| `date` | 发布日期 | `2026-01-10` |
| `tags` | 标签列表 | `[Go, 教程]` |
| `draft` | 草稿模式 | `true` / `false` |
| `pinned` | 置顶文章 | `true` / `false` |

---

## 分类管理

分类就是 `content/blog/` 下的文件夹：

```
content/blog/
├── default/     # 默认分类
├── tech/        # 技术分类
└── life/        # 生活分类
```

创建新文件夹即创建新分类。

---

## 独立页面

在 `content/page/` 下创建 `.md` 文件，会自动显示在导航栏：

```
content/page/
├── about.md     # 关于页面 → /page/about.html
└── links.md     # 友链页面 → /page/links.html
```

---

## 配置说明

`config.yaml` 主要配置项：

```yaml
site:
    title: 站点标题
    description: 站点描述
    author: 作者名
    
    # 功能开关
    comments_enabled: true    # 评论功能
    toc_enabled: true         # 文章目录
    reading_time_enabled: true # 阅读时间
    
    # 外观
    default_theme: auto       # light / dark / auto
    accent_color: "#2563eb"   # 主题色
    logo: ""                  # Logo 图片地址
    
    # 页脚
    footer:
        copyright: "© 2026 My Blog"
        icp: ""               # 备案号
        links:                # 社交链接
            - name: GitHub
              url: https://github.com/xxx
              icon: fa-brands fa-github
```

---

## 其他部署方式

### Docker 部署

```bash
docker-compose up -d
```

### 服务器部署

```bash
go build -o mdblog
./mdblog
```

配合 Nginx 反向代理使用。

### 静态部署（GitHub Pages）

如果不需要评论和后台功能，可以生成静态站点：

```bash
go run main.go --build
```

生成的文件在 `public/` 目录，可部署到 GitHub Pages。

---

## 目录结构

```
mdblog/
├── content/
│   ├── blog/        # 博客文章
│   └── page/        # 独立页面
├── data/            # 评论、统计数据
├── uploads/         # 上传的图片
├── themes/          # 主题文件
├── config.yaml      # 配置文件
└── main.go          # 程序入口
```

---

## 常见问题

### 如何修改端口？

编辑 `config.yaml` 中的 `server.port`

### 如何备份？

备份以下目录：
- `content/` - 文章
- `data/` - 评论和统计
- `uploads/` - 图片
- `config.yaml` - 配置

### 如何更换主题？

1. 在 `themes/` 下创建新主题
2. 修改 `config.yaml` 中的 `theme: 主题名`

---

有问题？欢迎在 [GitHub Issues](https://github.com/wdcbot/mdblog/issues) 反馈！
