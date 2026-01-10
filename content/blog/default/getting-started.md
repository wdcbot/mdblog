---
title: "mdblog 使用指南"
date: 2026-01-10
tags: [教程, 入门]
pinned: true
---

欢迎使用 mdblog！这是一个轻量级的 Markdown 博客系统，无需数据库，简单易用。

## 快速开始

### 1. 下载项目

```bash
git clone https://github.com/wdcbot/mdblog.git
cd mdblog
```

### 2. 配置

复制配置文件并修改：

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`，修改管理员密码和站点信息：

```yaml
server:
    port: 8080

admin:
    username: admin
    password: 你的密码
    jwt_secret: 随机字符串

site:
    title: 我的博客
    description: 博客描述
```

### 3. 运行

```bash
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

## 部署

### 本地运行

```bash
go build -o mdblog
./mdblog
```

### Docker 部署

```bash
docker-compose up -d
```

### 服务器部署

1. 编译：`go build -o mdblog`
2. 上传到服务器
3. 使用 systemd 或 supervisor 管理进程
4. 配置 Nginx 反向代理

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
