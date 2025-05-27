# 使用官方 Golang 镜像作为构建环境
FROM golang:1.24.2-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o mrs ./cmd/server/main.go

# 使用更小的基础镜像运行应用
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/mrs .

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./mrs"]
