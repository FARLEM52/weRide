FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o payment_service ./internal/services/payment_service/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/payment_service .
COPY --from=builder /app/internal/services/payment_service/config ./internal/services/payment_service/config
EXPOSE 50053
CMD ["./payment_service"]
