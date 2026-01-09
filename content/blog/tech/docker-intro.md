---
title: "Docker 容器化入门"
date: 2026-01-05
tags: [Docker, DevOps, 容器]
---

Docker 是一个开源的容器化平台，让开发者可以将应用及其依赖打包到一个可移植的容器中。

## 核心概念

- **镜像 (Image)** - 只读模板，包含运行应用所需的一切
- **容器 (Container)** - 镜像的运行实例
- **Dockerfile** - 构建镜像的脚本

## 常用命令

```bash
# 拉取镜像
docker pull nginx

# 运行容器
docker run -d -p 80:80 nginx

# 查看运行中的容器
docker ps

# 停止容器
docker stop <container_id>
```

## Dockerfile 示例

```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o main .
CMD ["./main"]
```

容器化让部署变得简单可重复。
