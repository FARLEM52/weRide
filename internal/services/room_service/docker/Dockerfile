FROM golang:1.23.5-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal/services/room_service/ /app/internal/services/room_service/
COPY internal/services/user_service/protoc/gen/go/ /app/internal/services/user_service/protoc/gen/go/
COPY internal/services/payment_service/pb/ /app/internal/services/payment_service/pb/
COPY internal/pkg/ /app/internal/pkg/

WORKDIR /app/internal/services/room_service
RUN CGO_ENABLED=0 GOOS=linux go build -o /room-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /room-service /app/
COPY --from=builder /app/internal/services/room_service/config /app/internal/services/room_service/config

EXPOSE 50051
CMD ["./room-service"]
