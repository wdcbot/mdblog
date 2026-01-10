---
title: "05. 青峰 Swagger 生产部署：Docker 与最佳实践"
date: 2026-01-10
tags: [青峰, Swagger, Go, Docker, 部署]
---

本文介绍如何在生产环境中部署青峰 Swagger，包括 Docker 部署和最佳实践。

---

## 嵌入 Swagger 文档

生产环境推荐使用 `embed` 将文档嵌入二进制文件：

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
        Title:   "我的 API",
        BasePath: "/doc",
        DocJSON: swaggerJSON,  // 使用嵌入的文档
    }))
    
    r.Run(":8080")
}
```

优点：
- 无需在服务器上维护 `docs/` 目录
- 部署更简单，只需一个二进制文件
- 文档版本与代码版本一致

## Dockerfile

### 单阶段构建（简单）

```dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .

RUN go build -o main .

EXPOSE 8080
CMD ["./main"]
```

### 多阶段构建（推荐）

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
```

优点：
- 镜像体积小（约 20MB vs 1GB+）
- 不包含编译工具，更安全

## docker-compose.yml

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
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## 构建和运行

```bash
# 构建镜像
docker build -t my-api .

# 运行容器
docker run -d -p 8080:8080 --name my-api my-api

# 使用 docker-compose
docker-compose up -d
```

## 生产环境配置

### 1. 禁用调试功能（可选）

如果不希望外部用户调试 API：

```go
qingfeng.Config{
    EnableDebug: false,
}
```

### 2. 设置正确的 Host

```go
// main.go 顶部注释
// @host api.example.com
// @BasePath /api/v1
```

### 3. 配置 HTTPS

使用 Nginx 反向代理：

```nginx
server {
    listen 443 ssl;
    server_name api.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 4. 访问控制

如果文档只对内部开放，可以添加认证：

```go
// 简单的 Basic Auth
authorized := r.Group("/doc", gin.BasicAuth(gin.Accounts{
    "admin": "password",
}))
authorized.GET("/*any", qingfeng.Handler(config))
```

## CI/CD 集成

### GitHub Actions

```yaml
name: Build and Deploy

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install swag
        run: go install github.com/swaggo/swag/cmd/swag@latest
      
      - name: Generate docs
        run: swag init
      
      - name: Build
        run: go build -o main .
      
      - name: Build Docker image
        run: docker build -t my-api .
      
      - name: Push to registry
        run: |
          docker tag my-api registry.example.com/my-api:${{ github.sha }}
          docker push registry.example.com/my-api:${{ github.sha }}
```

## 自动生成文档

开发环境可以启用自动生成：

```go
qingfeng.Config{
    AutoGenerate:  true,  // 启动时自动运行 swag init
    SwagSearchDir: ".",
    SwagOutputDir: "./docs",
}
```

> ⚠️ 生产环境建议关闭，使用 CI/CD 预先生成文档。

## 性能优化

### 1. 使用 Release 模式

```go
gin.SetMode(gin.ReleaseMode)
```

或设置环境变量：
```bash
export GIN_MODE=release
```

### 2. 启用 Gzip 压缩

```go
import "github.com/gin-contrib/gzip"

r.Use(gzip.Gzip(gzip.DefaultCompression))
```

### 3. 设置合理的超时

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      r,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
}
srv.ListenAndServe()
```

## 监控和日志

### 健康检查端点

```go
r.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})
```

### 请求日志

```go
r.Use(gin.Logger())
```

---

## 总结

生产部署清单：

- [ ] 使用 `embed` 嵌入文档
- [ ] 多阶段 Docker 构建
- [ ] 设置正确的 Host 和 BasePath
- [ ] 配置 HTTPS
- [ ] 考虑访问控制
- [ ] CI/CD 自动生成文档
- [ ] 启用 Release 模式
- [ ] 添加健康检查

---

**项目地址**: [github.com/wdcbot/qingfeng](https://github.com/wdcbot/qingfeng)
