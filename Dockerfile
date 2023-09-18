# 使用官方的 Golang 基础镜像
FROM golang:1.21 AS build

# 设置工作目录
WORKDIR /app
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
# 复制 go.mod 和 go.sum 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源代码到容器中
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o temp-container-manager .

# 使用轻量的 alpine 镜像作为最终镜像
FROM alpine:20230901

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件到最终镜像
COPY --from=build /app/temp-container-manager .

# 设置环境变量
ENV PORT=80

# 暴露端口
EXPOSE $PORT

# 启动应用程序
CMD ["./temp-container-manager", "/app/temp-container-manager/conf/config.yaml"]
