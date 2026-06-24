# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.24 AS builder

WORKDIR /src

# 依存解決（キャッシュ効率のため先に実行）
COPY go.mod go.sum ./
RUN go mod download

# ソースをコピーしてビルド
COPY . .
# migrations は embed.FS でバイナリに含まれるため別途コピー不要
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/api /app/api

# Cloud Run は PORT 環境変数を注入する（デフォルト 8080）
ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
