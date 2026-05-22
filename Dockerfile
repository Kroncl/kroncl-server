FROM golang:1.25-alpine AS builder

WORKDIR /app

# Устанавливаем git (необходим для go mod download)
RUN apk add --no-cache git

# Используем только российское зеркало или официальный прокси
ENV GOPROXY=https://goproxy.ru,https://proxy.golang.org,direct
ENV GOSUMDB=off
ENV GO111MODULE=on

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости с повторными попытками
RUN for i in 1 2 3 4 5; do \
      go mod download && break || \
      echo "Retry $i/5 failed, waiting 5 seconds..." && sleep 5; \
    done

# Копируем исходники
COPY . .

# Собираем основное приложение
RUN go build -o main cmd/api/main.go

# Собираем мигратор
RUN go build -o migrate cmd/migrate/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарники из билдера
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

# Копируем папку с миграциями и скрипты
COPY --from=builder /app/migrations ./
COPY docker-entrypoint.sh .

RUN sed -i 's/\r$//' docker-entrypoint.sh && \
    chmod +x docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./docker-entrypoint.sh"]