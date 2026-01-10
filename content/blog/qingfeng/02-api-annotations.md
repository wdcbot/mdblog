---
title: "02. 青峰 Swagger 注释详解：让你的 API 文档更完善"
date: 2026-01-10
tags: [青峰, Swagger, Go, 注释]
---

Swagger 注释是生成 API 文档的核心。本文详细介绍各种注释的用法。

---

## 主文件注释

在 `main.go` 顶部添加项目级别的注释：

```go
// @title 项目名称
// @version 1.0
// @description 项目描述，支持多行
// @description 第二行描述

// @contact.name 技术支持
// @contact.email support@example.com
// @contact.url https://example.com

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 Bearer {token}
```

## 接口注释

### 基本结构

```go
// @Summary 接口简介（一句话）
// @Description 详细描述（可多行）
// @Tags 分组标签
// @Accept json
// @Produce json
// @Param 参数名 位置 类型 是否必填 "描述"
// @Success 状态码 {类型} 数据结构 "描述"
// @Failure 状态码 {类型} 数据结构 "描述"
// @Router /路径 [方法]
// @Security BearerAuth
```

### 参数位置

| 位置 | 说明 | 示例 |
|------|------|------|
| `path` | URL 路径参数 | `/users/{id}` 中的 `id` |
| `query` | URL 查询参数 | `?page=1&size=10` |
| `header` | 请求头 | `Authorization` |
| `body` | 请求体 | JSON 数据 |
| `formData` | 表单数据 | 文件上传 |

## 实战示例

### 1. GET 请求 - 查询参数

```go
// @Summary 获取用户列表
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} Response{data=[]User}
// @Router /users [get]
func listUsers(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    size := c.DefaultQuery("size", "10")
    // ...
}
```

### 2. GET 请求 - 路径参数

```go
// @Summary 获取用户详情
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} Response{data=User}
// @Failure 404 {object} Response
// @Router /users/{id} [get]
func getUser(c *gin.Context) {
    id := c.Param("id")
    // ...
}
```

### 3. POST 请求 - JSON 请求体

```go
// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required" example:"张三"`
    Email    string `json:"email" binding:"required,email" example:"test@example.com"`
    Age      int    `json:"age" binding:"gte=0,lte=150" example:"25"`
    Password string `json:"password" binding:"required,min=6" example:"123456"`
}

// @Summary 创建用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "用户信息"
// @Success 201 {object} Response{data=User}
// @Failure 400 {object} Response
// @Router /users [post]
func createUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // ...
    }
}
```

### 4. PUT 请求 - 路径参数 + 请求体

```go
// @Summary 更新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Param user body UpdateUserRequest true "更新信息"
// @Success 200 {object} Response{data=User}
// @Router /users/{id} [put]
func updateUser(c *gin.Context) {
    id := c.Param("id")
    // ...
}
```

### 5. DELETE 请求

```go
// @Summary 删除用户
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /users/{id} [delete]
func deleteUser(c *gin.Context) {
    id := c.Param("id")
    // ...
}
```

### 6. 文件上传

```go
// @Summary 上传头像
// @Tags 用户管理
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "用户 ID"
// @Param avatar formData file true "头像文件"
// @Success 200 {object} Response{data=string}
// @Router /users/{id}/avatar [post]
func uploadAvatar(c *gin.Context) {
    file, _ := c.FormFile("avatar")
    // ...
}
```

### 7. 枚举参数

```go
// @Summary 获取订单列表
// @Tags 订单
// @Param status query string false "订单状态" Enums(pending, paid, shipped, completed)
// @Param sort query string false "排序方式" Enums(asc, desc) default(desc)
// @Router /orders [get]
func listOrders(c *gin.Context) {
    status := c.Query("status")
    // ...
}
```

## 数据结构定义

### 通用响应结构

```go
// Response 通用响应
type Response struct {
    Code    int         `json:"code" example:"0"`
    Message string      `json:"message" example:"success"`
    Data    interface{} `json:"data"`
}

// User 用户模型
type User struct {
    ID        int       `json:"id" example:"1"`
    Name      string    `json:"name" example:"张三"`
    Email     string    `json:"email" example:"test@example.com"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 分页响应

```go
// PageData 分页数据
type PageData struct {
    Total int         `json:"total" example:"100"`
    Page  int         `json:"page" example:"1"`
    Size  int         `json:"size" example:"10"`
    Items interface{} `json:"items"`
}

// @Success 200 {object} Response{data=PageData{items=[]User}}
```

## 多级标签分组

使用 `-` 分隔符创建层级：

```go
// @Tags Admin-用户管理
func adminListUsers() {}

// @Tags Admin-权限管理
func adminListRoles() {}

// @Tags Client-用户
func clientGetProfile() {}
```

生成的目录结构：
```
├── Admin
│   ├── 用户管理
│   └── 权限管理
└── Client
    └── 用户
```

---

## 生成文档

每次修改注释后，重新运行：

```bash
swag init
```

常用参数：
```bash
# 解析依赖包中的类型
swag init --parseDependency

# 解析内部包
swag init --parseInternal

# 指定输出目录
swag init -o ./docs
```

---

**下一篇**: [配置主题和样式](/qingfeng/03-themes.html)
