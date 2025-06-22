# --- Stage 1: Builder ---
# 使用官方的Go镜像作为构建环境
FROM golang:1.24.2-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache gcc musl-dev

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o seed ./cmd/seed/main.go

# --- Stage 2: Final Image ---
# 使用一个极简的基础镜像，比如 alpine
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata curl wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    rm -rf /var/cache/apk/*

# 设置工作目录
WORKDIR /app

# 创建必要的目录
RUN mkdir -p /app/config /app/var/log && \
    chown -R nobody:nobody /app && \
    chmod -R 755 /app

# 从 builder 阶段复制编译好的二进制文件到当前阶段
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .
COPY --from=builder /app/seed .

# 复制配置文件目录
# 这样我们可以在容器内访问 configs/app.yaml 等文件
COPY config ./config

# 设置权限
RUN chmod +x /app/server /app/migrate /app/seed && \
    chown -R nobody:nobody /app

# 切换到非 root 用户
USER nobody

# 暴露服务端口 (与你的app.yaml中的端口保持一致)
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 容器启动时运行的命令
CMD ["./server"]
