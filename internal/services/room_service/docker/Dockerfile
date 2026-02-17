# Этап сборки
FROM golang:1.23.5-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Копируем только необходимые части проекта
COPY internal/services/room_service/ /app/internal/services/room_service/
COPY internal/pkg/ /app/internal/pkg/

# Собираем приложение
WORKDIR /app/internal/services/room_service
RUN CGO_ENABLED=0 GOOS=linux go build -o /room-service ./cmd/main.go

# Этап запуска
FROM alpine:latest

# Создаем структуру директорий и копируем конфиг
RUN mkdir -p /app/config
COPY --from=builder /app/internal/services/room_service/config/config.yaml /app/config/

# Копируем бинарник
COPY --from=builder /room-service /app/

WORKDIR /app
EXPOSE 50051 8081
CMD ["./room-service", "--config", "/app/config/local.yaml"]