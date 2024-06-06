# 使用官方Go镜像作为构建基础镜像
FROM golang:1.22 AS builder

ENV GOPROXY=https://goproxy.io
# ENV GOPROXY=https://goproxy.cn
# 设置工作目录
WORKDIR /app

# 将go.mod和go.sum复制到工作目录
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 将项目文件复制到工作目录
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook .

# 使用一个较小的镜像作为运行时环境
FROM alpine:3.12

# 设置工作目录
WORKDIR /root/

# 将构建阶段的可执行文件复制到运行时镜像
COPY --from=builder /app/webhook .

# 将证书文件复制到运行时镜像
COPY tls.crt /tmp/ssl/tls.crt
COPY tls.key /tmp/ssl/tls.key

# 暴露端口
EXPOSE 8443

# 运行可执行文件
CMD ["./webhook"]
