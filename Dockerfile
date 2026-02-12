# ---- 1. Сборка зависимостей (локально, если есть vendor) ----
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем всё (включая vendor, если есть)
COPY . .

# Если папка vendor существует, используем её
# Если нет — пробуем скачать с явным прокси
RUN if [ ! -d "vendor" ]; then \
        go env -w GOPROXY=https://proxy.golang.org,direct; \
        go mod download; \
    fi

# Сборка с vendor (если есть) или обычная
RUN CGO_ENABLED=0 GOOS=linux go build \
    $( [ -d vendor ] && echo "-mod=vendor" ) \
    -v -o bank_api ./cmd/main.go

# ---- 2. Финальный образ ----
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bank_api .
COPY --from=builder /app/internal/database/postgres/schema.sql ./migrations/

EXPOSE 8080

CMD ["./bank_api"]