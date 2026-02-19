FROM golang:1.23.5-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY api/ /app/api/
COPY internal/pkg/ /app/internal/pkg/
COPY internal/services/user_service/protoc/gen/go/ /app/internal/services/user_service/protoc/gen/go/
COPY internal/services/room_service/pb/ /app/internal/services/room_service/pb/
COPY internal/services/payment_service/pb/ /app/internal/services/payment_service/pb/

WORKDIR /app/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /api-gateway /app/
COPY --from=builder /app/api/config /app/api/config

EXPOSE 8080
CMD ["./api-gateway"]
