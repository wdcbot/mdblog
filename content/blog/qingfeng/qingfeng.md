---
pinned: true
title: "青峰 Swagger (QingFeng) 完整使用指南"
date: 2026-01-09
tags: [教程, 文档,Swagger]
---
青峰 Swagger 是一个美观、强大的 Swagger UI 替代方案，支持 Gin、Fiber、Echo、Chi 等主流 Go Web 框架。为 Go 开发者提供更好的 API 文档体验。
-----------------------------------------------------------------------------------------------------------------------------------------

> 版本: v1.6.2 | 作者: wdc | 许可证: MIT

青峰 Swagger 是一个美观、强大的 Swagger UI 替代方案，支持 Gin、Fiber、Echo、Chi 等主流 Go Web 框架。为 Go 开发者提供更好的 API 文档体验。

---

## 目录

1. [项目简介](#1-项目简介)
2. [安装方式](#2-安装方式)
3. [快速开始](#3-快速开始)
4. [完整配置参数](#4-完整配置参数)
5. [UI 主题系统](#5-ui-主题系统)
6. [多框架支持](#6-多框架支持)
7. [高级功能](#7-高级功能)
8. [Swag 注释指南](#8-swag-注释指南)
9. [Docker 部署](#9-docker-部署)
10. [常见问题](#10-常见问题)
11. [更新日志](#11-更新日志)

---

## 1. 项目简介

### 1.1 核心特性


| 特性              | 说明                                               |
| ----------------- | -------------------------------------------------- |
| 🎨 多主题支持     | Default、Minimal、Modern 三种 UI 风格              |
| 🌓 深色/浅色模式  | 支持主题切换，保护眼睛                             |
| 🎯 多种主题色     | 蓝、绿、紫、橙、红、青六种主题色可选               |
| 🔍 快速搜索       | 实时搜索接口，快速定位（支持 Ctrl+K 快捷键）       |
| 🐛 在线调试       | 内置 API 调试工具，类似 Postman                    |
| 🔑 全局请求头     | 支持配置全局 Headers（如 Authorization）           |
| 🪄 Token 自动提取 | 从响应中自动提取 Token 设置到全局参数              |
| 🔄 自动生成文档   | 启动时自动运行 swag init                           |
| 📦 零依赖前端     | 使用 embed.FS 内嵌，无需额外部署                   |
| 🚀 简单集成       | 一行代码接入现有项目                               |
| 📱 移动端适配     | 完美支持手机访问，侧边栏抽屉式交互                 |
| 💾 设置持久化     | 主题、UI 风格、全局参数自动保存到本地              |
| ✨ JSON 语法高亮  | 响应结果彩色高亮显示                               |
| 🔌 多框架支持     | 原生支持 Gin，其他框架可通过标准 http.Handler 适配 |

### 1.2 项目地址

- **GitHub**: https://github.com/wdcbot/qingfeng
- **Gitee (国内镜像)**: https://gitee.com/xiaowan1997/qingfeng

---

## 2. 安装方式

### 2.1 使用 go get 安装

```bash
# GitHub (推荐)
go get github.com/wdcbot/qingfeng@latest

# Gitee 国内镜像
go get gitee.com/xiaowan1997/qingfeng@latest
```

### 2.2 安装 swag 工具

swag 是用于解析 Go 代码注释生成 Swagger 文档的工具：

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

验证安装：

```bash
swag --version
```

---

## 3. 快速开始

### 3.1 从零开始创建项目

```bash
# 1. 创建项目目录
mkdir myapi && cd myapi
go mod init myapi

# 2. 安装依赖
go get github.com/gin-gonic/gin
go get github.com/wdcbot/qingfeng@latest
```

### 3.2 创建 main.go

```go
package main

import (
    "github.com/gin-gonic/gin"
    qingfeng "github.com/wdcbot/qingfeng"
)

// @title 我的 API
// @version 1.0
// @description 这是我的第一个 API
// @host localhost:8080
// @BasePath /api

func main() {
    r := gin.Default()

    // 注册文档 UI
    r.GET("/doc/*any", qingfeng.Handler(qingfeng.Config{
        Title:    "我的 API 文档",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    }))

    // API 路由
    r.GET("/api/hello", hello)

    r.Run(":8080")
}

// @Summary 打招呼
// @Tags 示例
// @Success 200 {string} string "Hello World"
// @Router /hello [get]
func hello(c *gin.Context) {
    c.JSON(200, gin.H{"message": "Hello World"})
}
```

### 3.3 生成文档并运行

```bash
# 生成 swagger 文档
swag init

# 运行项目
go run main.go
```

### 3.4 访问文档

打开浏览器访问：http://localhost:8080/doc/

---

## 4. 完整配置参数

### 4.1 Config 结构体

```go
type Config struct {
    // 基础配置
    Title         string        // 文档标题，默认 "API Documentation"
    Description   string        // 文档描述
    Version       string        // API 版本号，默认 "1.0.0"
    BasePath      string        // 文档路由前缀，默认 "/doc"
  
    // 文档来源（二选一）
    DocPath       string        // swagger.json 文件路径，默认 "./docs/swagger.json"
    DocJSON       []byte        // 直接传入 swagger JSON 内容
  
    // 功能开关
    EnableDebug   bool          // 是否启用在线调试，默认 true
    DarkMode      bool          // 是否默认深色模式，默认 false
    PersistParams *bool         // 是否保存调试参数到 sessionStorage，默认 true
  
    // UI 配置
    UITheme       UITheme       // UI 主题风格，默认 ThemeDefault
    Logo          string        // 自定义 Logo URL 或 base64
    LogoLink      string        // Logo 点击跳转链接
  
    // 全局请求头
    GlobalHeaders []Header      // 全局请求头配置
  
    // 自动生成文档
    AutoGenerate  bool          // 启动时自动运行 swag init，默认 false
    SwagSearchDir string        // swag 搜索目录，默认 "."
    SwagOutputDir string        // swagger 文件输出目录，默认 "./docs"
    SwagArgs      []string      // swag init 的额外参数
  
    // 多环境配置
    Environments  []Environment // 环境配置列表
}
```

### 4.2 Header 结构体

```go
type Header struct {
    Key   string `json:"key"`   // 请求头名称，如 "Authorization"
    Value string `json:"value"` // 请求头值，如 "Bearer xxx"
}
```

### 4.3 Environment 结构体

```go
type Environment struct {
    Name    string `json:"name"`    // 环境名称，如 "本地开发"
    BaseURL string `json:"baseUrl"` // API 基础 URL
}
```

### 4.4 UITheme 常量

```go
const (
    ThemeDefault UITheme = "default"  // 默认主题 - 经典蓝色风格
    ThemeMinimal UITheme = "minimal"  // 简约主题 - 黑白极简
    ThemeModern  UITheme = "modern"   // 现代主题 - 渐变毛玻璃
)
```

### 4.5 完整配置示例

```go
r.GET("/doc/*any", qingfeng.Handler(qingfeng.Config{
    // 基础信息
    Title:       "我的 API",
    Description: "API 文档描述",
    Version:     "1.0.0",
    BasePath:    "/doc",
    DocPath:     "./docs/swagger.json",
  
    // 功能配置
    EnableDebug: true,
    DarkMode:    false,
  
    // UI 主题
    UITheme: qingfeng.ThemeDefault,
  
    // 自定义 Logo
    Logo:     "https://example.com/logo.png",
    LogoLink: "https://example.com",
  
    // 全局请求头
    GlobalHeaders: []qingfeng.Header{
        {Key: "Authorization", Value: "Bearer your-token"},
        {Key: "X-API-Key", Value: "your-api-key"},
    },
  
    // 自动生成文档
    AutoGenerate:  true,
    SwagSearchDir: ".",
    SwagOutputDir: "./docs",
    SwagArgs:      []string{"--parseDependency", "--parseInternal"},
  
    // 多环境配置
    Environments: []qingfeng.Environment{
        {Name: "本地开发", BaseURL: "/api/v1"},
        {Name: "测试环境", BaseURL: "https://test-api.example.com/api/v1"},
        {Name: "生产环境", BaseURL: "https://api.example.com/api/v1"},
    },
}))
```

### 4.6 配置参数详细说明


| 参数          | 类型          | 默认值                | 说明                                            |
| ------------- | ------------- | --------------------- | ----------------------------------------------- |
| Title         | string        | "API Documentation"   | 文档标题，显示在页面顶部                        |
| Description   | string        | ""                    | 文档描述信息                                    |
| Version       | string        | "1.0.0"               | API 版本号                                      |
| BasePath      | string        | "/doc"                | 文档路由前缀，访问路径为`{BasePath}/`           |
| DocPath       | string        | "./docs/swagger.json" | swagger.json 文件路径                           |
| DocJSON       | []byte        | nil                   | 直接传入 swagger JSON 内容（与 DocPath 二选一） |
| EnableDebug   | bool          | true                  | 是否启用在线调试功能                            |
| DarkMode      | bool          | false                 | 是否默认使用深色模式                            |
| PersistParams | *bool         | nil (默认 true)       | 是否将调试参数保存到 sessionStorage             |
| UITheme       | UITheme       | ThemeDefault          | UI 主题风格                                     |
| GlobalHeaders | []Header      | nil                   | 全局请求头，会自动添加到所有 API 请求           |
| AutoGenerate  | bool          | false                 | 启动时是否自动运行 swag init                    |
| SwagSearchDir | string        | "."                   | swag 搜索目录                                   |
| SwagOutputDir | string        | "./docs"              | swagger 文件输出目录                            |
| SwagArgs      | []string      | nil                   | swag init 的额外参数                            |
| Logo          | string        | ""                    | 自定义 Logo URL 或 base64 编码                  |
| LogoLink      | string        | ""                    | Logo 点击跳转链接                               |
| Environments  | []Environment | nil                   | 多环境配置列表                                  |

---

## 5. UI 主题系统

### 5.1 三种主题风格


| 主题    | 常量                    | 特点                                   |
| ------- | ----------------------- | -------------------------------------- |
| Default | `qingfeng.ThemeDefault` | 经典蓝色风格，功能完整，适合大多数场景 |
| Minimal | `qingfeng.ThemeMinimal` | 黑白极简，专业干净，适合正式文档       |
| Modern  | `qingfeng.ThemeModern`  | 渐变毛玻璃，视觉冲击，适合展示         |

### 5.2 主题切换方式

**方式一：代码配置**

```go
qingfeng.Config{
    UITheme: qingfeng.ThemeModern,
}
```

**方式二：URL 参数**

```
http://localhost:8080/doc/?theme=modern
http://localhost:8080/doc/?theme=minimal
http://localhost:8080/doc/?theme=default
```

**方式三：界面切换**

点击页面右上角的主题切换按钮，选择喜欢的主题。

### 5.3 主题色配置

支持 6 种主题色：蓝、绿、紫、橙、红、青

在界面中点击「主题」按钮可以切换主题色，设置会自动保存到浏览器本地存储。

### 5.4 深色/浅色模式

- 点击页面右上角的 🌙/☀️ 图标切换
- 或通过配置 `DarkMode: true` 设置默认深色模式
- 用户切换后的设置会保存到本地存储

---

## 6. 多框架支持

青峰 Swagger 提供标准 `http.Handler`，可适配任何 Go Web 框架。

### 6.1 Gin (原生支持)

```go
import (
    "github.com/gin-gonic/gin"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    r := gin.Default()
  
    r.GET("/doc/*any", qingfeng.Handler(qingfeng.Config{
        Title:    "我的 API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    }))
  
    r.Run(":8080")
}
```

### 6.2 Fiber

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/adaptor"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    app := fiber.New()
  
    app.Use("/doc", adaptor.HTTPHandler(qingfeng.HTTPHandler(qingfeng.Config{
        Title:    "我的 API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    })))
  
    app.Listen(":8080")
}
```

### 6.3 Echo

```go
import (
    "github.com/labstack/echo/v4"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    e := echo.New()
  
    e.GET("/doc/*", echo.WrapHandler(qingfeng.HTTPHandler(qingfeng.Config{
        Title:    "我的 API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    })))
  
    e.Start(":8080")
}
```

### 6.4 Chi

```go
import (
    "net/http"
    "github.com/go-chi/chi/v5"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    r := chi.NewRouter()
  
    r.Handle("/doc/*", qingfeng.HTTPHandler(qingfeng.Config{
        Title:    "我的 API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    }))
  
    http.ListenAndServe(":8080", r)
}
```

### 6.5 标准库 net/http

```go
import (
    "net/http"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    http.Handle("/doc/", qingfeng.HTTPHandler(qingfeng.Config{
        Title:    "我的 API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    }))
  
    http.ListenAndServe(":8080", nil)
}
```

---

## 7. 高级功能

### 7.1 全局请求头

预设全局请求头，会自动添加到所有 API 请求中：

```go
qingfeng.Config{
    GlobalHeaders: []qingfeng.Header{
        {Key: "Authorization", Value: "Bearer your-token"},
        {Key: "X-API-Key", Value: "your-api-key"},
        {Key: "X-Request-ID", Value: "unique-id"},
    },
}
```

也可以在界面中通过「全局参数」按钮动态配置。

### 7.2 Token 自动提取

在界面中配置 Token 提取规则，可以从 API 响应中自动提取 Token 并设置到全局参数：

1. 点击「Token」按钮
2. 添加提取规则：
   - 响应字段路径：如 `data.token`
   - 目标 Header：如 `Authorization`
   - 前缀：如 `Bearer `

### 7.3 多环境配置

配置多个环境，方便在开发、测试、生产环境间切换：

```go
qingfeng.Config{
    Environments: []qingfeng.Environment{
        {Name: "本地开发", BaseURL: "http://localhost:8080/api/v1"},
        {Name: "测试环境", BaseURL: "https://test-api.example.com/api/v1"},
        {Name: "生产环境", BaseURL: "https://api.example.com/api/v1"},
    },
}
```

在界面顶部会显示环境选择器，可以一键切换。

### 7.4 自定义 Logo

```go
qingfeng.Config{
    // 使用 URL
    Logo:     "https://example.com/logo.png",
    LogoLink: "https://example.com",
  
    // 或使用 base64
    // Logo: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
}
```

### 7.5 请求体模板

在调试面板中，可以将常用的请求体保存为模板：

1. 在请求体输入框上方点击「保存模板」
2. 输入模板名称
3. 下次使用时点击「模板」按钮选择已保存的模板

模板按接口保存，每个接口可以有多个模板。

### 7.6 自动生成文档

启用 `AutoGenerate` 后，每次启动服务会自动运行 `swag init`：

```go
qingfeng.Config{
    AutoGenerate:  true,
    SwagSearchDir: ".",
    SwagOutputDir: "./docs",
    SwagArgs:      []string{"--parseDependency", "--parseInternal"},
}
```

### 7.7 参数启用/禁用

每个参数前有勾选框，可控制是否发送该参数：

- 禁用的参数显示半透明，输入框禁用
- 勾选状态保存到 sessionStorage
- cURL 生成也会跳过禁用的参数

### 7.8 快捷键


| 快捷键             | 功能       |
| ------------------ | ---------- |
| `Ctrl/Cmd + K`     | 聚焦搜索框 |
| `Ctrl/Cmd + Enter` | 发送请求   |
| `Escape`           | 关闭弹窗   |

### 7.9 复制 cURL

在调试面板中，点击「复制 cURL」按钮可以复制当前请求的 cURL 命令，方便在终端中调试。

---

## 8. Swag 注释指南

### 8.1 主文件注释

在 main.go 或入口文件顶部添加：

```go
// @title API 标题
// @version 1.0
// @description API 描述信息
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
```

### 8.2 接口注释

```go
// @Summary 接口摘要
// @Description 接口详细描述
// @Tags 标签名
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param Authorization header string true "Bearer Token"
// @Param user body User true "用户信息"
// @Success 200 {object} Response{data=User}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /users/{id} [get]
// @Security ApiKeyAuth
func getUser(c *gin.Context) {
    // ...
}
```

### 8.3 参数类型说明


| 参数位置 | 说明         | 示例                                              |
| -------- | ------------ | ------------------------------------------------- |
| path     | URL 路径参数 | `@Param id path int true "用户ID"`                |
| query    | URL 查询参数 | `@Param page query int false "页码"`              |
| header   | 请求头参数   | `@Param Authorization header string true "Token"` |
| body     | 请求体参数   | `@Param user body User true "用户信息"`           |
| formData | 表单参数     | `@Param file formData file true "文件"`           |

### 8.4 多级目录（Tag 分组）

使用 `-` 分隔符创建多级目录：

```go
// @Tags Admin-User
func getUsers() {}

// @Tags Admin-Auth
func login() {}

// @Tags Public-Info
func getInfo() {}
```

这会生成如下目录结构：

```
├── Admin
│   ├── User
│   └── Auth
└── Public
    └── Info
```

### 8.5 枚举参数

```go
// @Param status query string true "状态" Enums(active, inactive, pending)
// @Param type query int true "类型" Enums(1, 2, 3)
```

### 8.6 文件上传

```go
// @Summary 上传文件
// @Accept multipart/form-data
// @Param file formData file true "文件"
// @Param user_id formData int true "用户ID"
// @Router /upload [post]
```

### 8.7 生成文档命令

```bash
# 基本用法
swag init

# 指定搜索目录
swag init -d ./cmd/api

# 指定输出目录
swag init -o ./docs

# 解析依赖
swag init --parseDependency --parseInternal

# 完整示例
swag init -d . -o ./docs --parseDependency --parseInternal
```

---

## 9. Docker 部署

### 9.1 使用 embed 嵌入文档（推荐）

```go
package main

import (
    "embed"
    "github.com/gin-gonic/gin"
    qingfeng "github.com/wdcbot/qingfeng"
)

//go:embed docs/swagger.json
var swaggerJSON []byte

func main() {
    r := gin.Default()
  
    r.GET("/doc/*any", qingfeng.Handler(qingfeng.Config{
        Title:    "我的 API",
        BasePath: "/doc",
        DocJSON:  swaggerJSON,  // 直接嵌入，无需 DocPath
    }))
  
    r.Run(":8080")
}
```

### 9.2 Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

# 运行阶段
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
# 不需要 COPY docs 目录！
EXPOSE 8080
CMD ["./main"]
```

### 9.3 docker-compose.yml

```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
    restart: unless-stopped
```

---

## 10. 常见问题

### 10.1 swagger.json 加载失败

**问题**：页面显示 "加载失败" 或 "swagger.json not found"

**解决方案**：

1. 确保已运行 `swag init` 生成文档
2. 检查 `DocPath` 配置是否正确
3. 确保 `docs/swagger.json` 文件存在

### 10.2 swag 命令未找到

**问题**：运行 `swag init` 提示命令未找到

**解决方案**：

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

确保 `$GOPATH/bin` 在系统 PATH 中。

### 10.3 接口不显示

**问题**：代码中有接口但文档中不显示

**解决方案**：

1. 检查是否添加了 swag 注释
2. 确保注释格式正确
3. 重新运行 `swag init`

### 10.4 深色模式不生效

**问题**：配置了 `DarkMode: true` 但页面还是浅色

**解决方案**：
清除浏览器本地存储，或在界面中手动切换主题。用户的手动设置优先级高于配置。

### 10.5 跨域问题

**问题**：调试时出现 CORS 错误

**解决方案**：
在后端添加 CORS 中间件：

```go
import "github.com/gin-contrib/cors"

r.Use(cors.Default())
```

### 10.6 文件上传不工作

**问题**：文件上传接口无法选择文件

**解决方案**：
确保参数注释正确：

```go
// @Param file formData file true "文件"
```

---

## 11. 更新日志

### v1.6.2 (2026-01-10)

- 修复暗黑模式文字看不清问题
- 简约主题添加主题切换功能
- 修复暗黑模式划词后看不清
- 更多按钮适配暗黑模式
- 更新 GitHub 链接地址
- 升级前端依赖版本

### v1.6.1 (2026-01-06)

- 新增 `PersistParams` 配置项
- 参数启用/禁用勾选功能
- 新增 `/doc.json` 路径支持
- 修复枚举参数默认值显示问题
- 修复表单模式布尔值类型问题

### v1.5.5 (2024-12-30)

- 多框架支持 (Fiber/Echo/Chi/标准库)
- 新增 `HTTPHandler()` 返回标准 `http.Handler`

### v1.5.0 (2024-12-26)

- 离线模式支持
- Tailwind CSS 和 Font Awesome 打包到二进制

### v1.4.2 (2024-12-25)

- 文件上传支持
- FormData 请求自动检测

### v1.4.0 (2024-12-24)

- 响应结构展示
- 请求体结构化展示
- 自定义 swag 参数
- 多级目录支持

### v1.3.0 (2024-12-22)

- 多环境支持
- 请求体模板
- 自定义 Logo
- 复制 cURL
- 快捷键支持

### v1.2.0 (2024-12-21)

- 移动端适配
- 调试数据持久化
- JSON 语法高亮

### v1.1.0 (2024-12-20)

- 多主题支持
- 深色模式
- Token 自动提取
- 全局请求头

### v1.0.0 (2024-12-19)

- 初始版本发布

---

## 联系方式

- **GitHub Issues**: https://github.com/wdcbot/qingfeng/issues
- **Gitee Issues**: https://gitee.com/xiaowan1997/qingfeng/issues

---

**青峰 Swagger** - 为 Go 开发者提供更好的 API 文档体验 ⚡️
