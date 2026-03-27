# 构建阶段
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=1 go build -o smtp-lite ./cmd/server/main.go

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 复制构建产物
COPY --from=builder /app/smtp-lite .
COPY --from=builder /app/frontend/dist ./frontend/dist

# 创建数据目录
RUN mkdir -p /data

# 暴露端口
EXPOSE 8090

# 运行
CMD ["./smtp-lite"]