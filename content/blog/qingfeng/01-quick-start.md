---
title: "01. 青峰 Swagger 快速入门：5 分钟搭建 API 文档"
date: 2026-01-10
tags: [青峰, Swagger, Go, 入门教程]
---

本教程带你从零开始，5 分钟内搭建一个带有漂亮 API 文档的 Go 项目。

---

## 前置条件

- Go 1.18 或更高版本
- 基本的 Go 语言知识

## 第一步：创建项目

```bash
# 创建项目目录
mkdir my-api
cd my-api

# 初始化 Go 模块
go mod init my-api
```

## 第二步：安装依赖

```bash
# 安装 Gin 框架
go get github.com/gin-gonic/gin

# 安装青峰 Swagger
go get github.com/wdcbot/qingfeng@latest

# 安装 swag 工具（用于生成文档）
go install github.com/swaggo/swag/cmd/swag@latest
```

## 第三步：编写代码

创建 `main.go`：

```go
package main

import (
    "github.com/gin-gonic/gin"
    qingfeng "github.com/wdcbot/qingfeng"
)

// @title 我的第一个 API
// @version 1.0
// @description 这是一个简单的示例 API
// @host localhost:8080
// @BasePath /api

func main() {
    r := gin.Default()

    // 注册 API 文档
    r.GET("/doc/*any", qingfeng.Handler(qingfeng.Config{
        Title:    "我的 API 文档",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    }))

    // API 路由
    api := r.Group("/api")
    {
        api.GET("/hello", hello)
        api.GET("/users/:id", getUser)
    }

    r.Run(":8080")
}

// @Summary 打招呼
// @Description 返回一个简单的问候语
// @Tags 示例
// @Produce json
// @Success 200 {object} map[string]string
// @Router /hello [get]
func hello(c *gin.Context) {
    c.JSON(200, gin.H{
        "message": "Hello, World!",
    })
}

// @Summary 获取用户
// @Description 根据 ID 获取用户信息
// @Tags 用户
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} map[string]interface{}
// @Router /users/{id} [get]
func getUser(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{
        "id":   id,
        "name": "张三",
        "email": "zhangsan@example.com",
    })
}
```

## 第四步：生成文档

```bash
swag init
```

这会在 `docs/` 目录下生成 `swagger.json` 文件。

## 第五步：运行项目

```bash
go run main.go
```

## 第六步：访问文档

打开浏览器访问：http://localhost:8080/doc/

你会看到一个漂亮的 API 文档界面！

---

## 下一步

- [添加更多接口和参数](/qingfeng/02-api-annotations.html)
- [配置主题和样式](/qingfeng/03-themes.html)
- [在线调试 API](/qingfeng/04-debugging.html)

---

**项目地址**: [github.com/wdcbot/qingfeng](https://github.com/wdcbot/qingfeng)
