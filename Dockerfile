# ---------- Frontend builder ----------
FROM node:24-alpine AS frontend-builder

# Alpine 换阿里源
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.aliyun.com|g' /etc/apk/repositories

WORKDIR /frontend

# Yarn 国内源
RUN corepack enable \
    && yarn config set registry https://registry.npmmirror.com

COPY frontend/package.json frontend/yarn.lock ./

RUN yarn install --frozen-lockfile

COPY frontend/ ./

RUN yarn build


# ---------- Go builder ----------
FROM golang:1.26-alpine AS builder

# Alpine 换阿里源
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.aliyun.com|g' /etc/apk/repositories \
    && apk add --no-cache \
       git \
       gcc \
       g++ \
       musl-dev \
       make

WORKDIR /src

# Go 国内代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# CGO
ENV CGO_ENABLED=1

# 依赖
COPY go.mod go.sum ./

RUN go mod download

# 代码
COPY . .

# 前端产物
COPY --from=frontend-builder /frontend/dist ./web/dist

# 构建
RUN GOOS=linux GOARCH=amd64 \
    go build \
    -tags embed_frontend \
    -ldflags="-s -w" \
    -o /app/server \
    ./cmd/server


# ---------- Runtime ----------
FROM alpine:3.20

# Alpine 换阿里源
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.aliyun.com|g' /etc/apk/repositories \
    && apk add --no-cache \
       ca-certificates \
       tzdata \
       sqlite-libs

WORKDIR /app

RUN adduser -D -h /app appuser

COPY --from=builder /app/server /app/server
COPY config.example.yaml /app/config.example.yaml

RUN mkdir -p /app/data/records \
    && chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/server"]