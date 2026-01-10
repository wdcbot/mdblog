---
title: "06. 青峰 Swagger 多框架集成：Gin、Fiber、Echo、Chi"
date: 2026-01-10
tags: [青峰, Swagger, Go, Gin, Fiber, Echo, Chi]
---

青峰 Swagger 支持多种 Go Web 框架，本文介绍各框架的集成方式。

---

## 框架支持列表

| 框架 | 集成方式 | 复杂度 |
|------|----------|--------|
| Gin | 原生支持 | ⭐ |
| Fiber | 适配器 | ⭐⭐ |
| Echo | 适配器 | ⭐⭐ |
| Chi | 标准 Handler | ⭐ |
| net/http | 标准 Handler | ⭐ |

## Gin（原生支持）

```go
package main

import (
    "github.com/gin-gonic/gin"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    r := gin.Default()
    
    r.GET("/doc/*any", qingfeng.Handler(qingfeng.Config{
        Title:    "Gin API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    }))
    
    r.Run(":8080")
}
```

## Fiber

Fiber 需要使用适配器将 `http.Handler` 转换：

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/adaptor"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    app := fiber.New()
    
    // 使用 adaptor 转换
    app.Use("/doc", adaptor.HTTPHandler(qingfeng.HTTPHandler(qingfeng.Config{
        Title:    "Fiber API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    })))
    
    // API 路由
    app.Get("/api/hello", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Hello from Fiber!"})
    })
    
    app.Listen(":8080")
}
```

### Fiber 注释示例

```go
// @Summary 打招呼
// @Tags 示例
// @Success 200 {object} map[string]string
// @Router /hello [get]
func hello(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "Hello!"})
}
```

## Echo

```go
package main

import (
    "github.com/labstack/echo/v4"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    e := echo.New()
    
    // 使用 WrapHandler 转换
    e.GET("/doc/*", echo.WrapHandler(qingfeng.HTTPHandler(qingfeng.Config{
        Title:    "Echo API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    })))
    
    // API 路由
    e.GET("/api/hello", hello)
    
    e.Start(":8080")
}

// @Summary 打招呼
// @Tags 示例
// @Success 200 {object} map[string]string
// @Router /hello [get]
func hello(c echo.Context) error {
    return c.JSON(200, map[string]string{"message": "Hello from Echo!"})
}
```

## Chi

Chi 原生支持 `http.Handler`：

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    
    // 直接使用 HTTPHandler
    r.Handle("/doc/*", qingfeng.HTTPHandler(qingfeng.Config{
        Title:    "Chi API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    }))
    
    // API 路由
    r.Get("/api/hello", hello)
    
    http.ListenAndServe(":8080", r)
}

// @Summary 打招呼
// @Tags 示例
// @Success 200 {object} map[string]string
// @Router /hello [get]
func hello(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"message": "Hello from Chi!"}`))
}
```

## 标准库 net/http

```go
package main

import (
    "encoding/json"
    "net/http"
    qingfeng "github.com/wdcbot/qingfeng"
)

func main() {
    // 文档路由
    http.Handle("/doc/", qingfeng.HTTPHandler(qingfeng.Config{
        Title:    "Standard Library API",
        BasePath: "/doc",
        DocPath:  "./docs/swagger.json",
    }))
    
    // API 路由
    http.HandleFunc("/api/hello", hello)
    
    http.ListenAndServe(":8080", nil)
}

// @Summary 打招呼
// @Tags 示例
// @Success 200 {object} map[string]string
// @Router /hello [get]
func hello(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Hello!"})
}
```

## Handler vs HTTPHandler

青峰提供两个函数：

| 函数 | 返回类型 | 适用框架 |
|------|----------|----------|
| `Handler()` | `gin.HandlerFunc` | Gin |
| `HTTPHandler()` | `http.Handler` | 其他所有框架 |

```go
// Gin 专用
qingfeng.Handler(config)

// 通用（标准库兼容）
qingfeng.HTTPHandler(config)
```

## 注意事项

### 1. BasePath 配置

确保 `BasePath` 与路由前缀一致：

```go
// ✅ 正确
r.Handle("/doc/*", qingfeng.HTTPHandler(qingfeng.Config{
    BasePath: "/doc",
}))

// ❌ 错误 - BasePath 不匹配
r.Handle("/api-docs/*", qingfeng.HTTPHandler(qingfeng.Config{
    BasePath: "/doc",  // 应该是 "/api-docs"
}))
```

### 2. 通配符路由

不同框架的通配符语法不同：

| 框架 | 语法 |
|------|------|
| Gin | `/doc/*any` |
| Fiber | `/doc/*` 或 `/doc` (Use) |
| Echo | `/doc/*` |
| Chi | `/doc/*` |
| net/http | `/doc/` |

### 3. CORS 配置

如果前端和 API 不同源，需要配置 CORS：

**Gin:**
```go
import "github.com/gin-contrib/cors"
r.Use(cors.Default())
```

**Fiber:**
```go
import "github.com/gofiber/fiber/v2/middleware/cors"
app.Use(cors.New())
```

**Echo:**
```go
import "github.com/labstack/echo/v4/middleware"
e.Use(middleware.CORS())
```

---

## 完整示例项目

查看 GitHub 仓库中的示例：

- [Gin 示例](https://github.com/wdcbot/qingfeng/tree/main/examples/gin)
- [Fiber 示例](https://github.com/wdcbot/qingfeng/tree/main/examples/fiber)
- [Echo 示例](https://github.com/wdcbot/qingfeng/tree/main/examples/echo)
- [Chi 示例](https://github.com/wdcbot/qingfeng/tree/main/examples/chi)

---

**项目地址**: [github.com/wdcbot/qingfeng](https://github.com/wdcbot/qingfeng)
