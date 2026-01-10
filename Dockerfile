# 构建阶段
FROM golang:1.25-alpine AS builder
LABEL "language"="go"

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mdblog .

# 运行阶段
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Shanghai

# 复制所有文件
COPY --from=builder /build/mdblog .
COPY --from=builder /build/config.yaml .
COPY --from=builder /build/themes/ ./themes/
COPY --from=builder /build/admin/ ./admin/
COPY --from=builder /build/docs/ ./docs/
COPY --from=builder /build/entrypoint.sh .

# GitHub 数据先复制到临时目录，启动时合并到 Volume
COPY --from=builder /build/content/ ./github-content/
COPY --from=builder /build/data/ ./github-data/
COPY --from=builder /build/uploads/ ./github-uploads/

# 创建挂载点目录
RUN mkdir -p /app/content /app/data /app/uploads && chmod +x /app/entrypoint.sh

EXPOSE 8080

CMD ["/app/entrypoint.sh"]
