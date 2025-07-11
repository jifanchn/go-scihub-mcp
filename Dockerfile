# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要工具
RUN apk add --no-cache git

# 设置Go代理为direct模式（直接连接，绕过所有代理）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
ENV CGO_ENABLED=0

# 复制 go mod 文件
COPY go.mod go.sum ./

# 禁用Git代理并下载依赖
RUN git config --global http.proxy "" && \
    git config --global https.proxy "" && \
    unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy ALL_PROXY all_proxy && \
    go mod download

# 复制源代码
COPY . .

# 构建应用
RUN GOOS=linux go build -ldflags="-w -s" -o scihub-mcp ./cmd/scihub-mcp

# 运行阶段
FROM alpine:latest

# 安装证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S scihub && \
    adduser -u 1001 -S scihub -G scihub

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/scihub-mcp .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建缓存目录
RUN mkdir -p ./cache && chown -R scihub:scihub /app

# 切换到非root用户
USER scihub

# 暴露端口
EXPOSE 8088

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8088/health || exit 1

# 启动应用
CMD ["./scihub-mcp", "--mcp-host", "0.0.0.0", "--mcp-port", "8088", "--proxy-enabled", "--proxy-host", "host.docker.internal", "--proxy-port", "3080", "mcp"] 