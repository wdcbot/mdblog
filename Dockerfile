# 构建阶段
# FROM golang:1.21-alpine AS builder
FROM golang:1.25-alpine AS builder
LABEL "language"="go"

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mdblog .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装 ca-certificates（用于 HTTPS 请求）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /build/mdblog .

# 复制必要文件
COPY config.yaml .
COPY themes/ ./themes/
COPY admin/ ./admin/

# 创建数据目录
RUN mkdir -p content/blog/default content/page data uploads

# 暴露端口
EXPOSE 8080

# 启动
CMD ["./mdblog"]
