# Этап сборки
FROM golang:1.23.5-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Копируем только необходимые части проекта
COPY internal/services/user_service/ /app/internal/services/user_service/
COPY internal/pkg/ /app/internal/pkg/

# Собираем приложение
WORKDIR /app/internal/services/user_service
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/main.go

# Этап запуска
FROM alpine:latest

# Создаем структуру директорий и копируем конфиг
RUN mkdir -p /app/config
COPY --from=builder /app/internal/services/user_service/config/local.yaml /app/config/

# Копируем бинарник
COPY --from=builder /user-service /app/

WORKDIR /app
EXPOSE 50052 8082
CMD ["./user-service", "--config", "/app/config/local.yaml"]