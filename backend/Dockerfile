# Build stage
ARG GO_VERSION=1.25.5
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# go modファイルをコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# APIバイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

# Workerバイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker ./cmd/worker

# Runtime stage
FROM alpine:3.23

WORKDIR /app

# HTTPS通信用のCA証明書をインストール
RUN apk --no-cache add ca-certificates

# 非rootユーザーを作成
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# ビルダーからバイナリをコピー
COPY --from=builder --chown=appuser:appuser /app/api .
COPY --from=builder --chown=appuser:appuser /app/worker .

# 非rootユーザーに切り替え
USER appuser

# Cloud RunはPORT環境変数を使用（デフォルト: 8080）
ENV PORT=8080

# ポートを公開
EXPOSE 8080

# APIを起動
CMD ["./api"]
