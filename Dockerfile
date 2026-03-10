# ── Stage 1: builder ──────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Зависимости для CGO (нужны некоторым пакетам)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Кешируем зависимости отдельным слоем
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники и собираем бинарь
COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w \
      -X main.Version=${VERSION} \
      -X main.GitCommit=${COMMIT} \
      -X main.BuildTime=${BUILD_TIME}" \
    -o /app/bin/api \
    ./cmd/api

# ── Stage 2: final ────────────────────────────────────────────────────────────
FROM scratch

# Копируем нужное из builder-а
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/bin/api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
