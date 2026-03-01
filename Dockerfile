FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем основное приложение
RUN go build -o main cmd/api/main.go

# Собираем мигратор
RUN go build -o migrate cmd/migrate/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Устанавливаем bash и netcat (для wait-for-it)
RUN apk add --no-cache bash netcat-openbsd

# Копируем бинарники из билдера
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

# Копируем папку с миграциями и скрипты
COPY --from=builder /app/migrations ./migrations
COPY wait-for-it.sh .
COPY docker-entrypoint.sh .

# Делаем скрипты исполняемыми
RUN chmod +x wait-for-it.sh docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./docker-entrypoint.sh"]