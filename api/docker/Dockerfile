FROM golang:1.23.5-alpine AS builder


RUN apk add --no-cache git


WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download


COPY api/ /app/api/
COPY internal/pkg/ /app/internal/pkg/
COPY internal/services/user_service/pb/ /app/internal/services/user_service/pb/
COPY internal/services/room_service/pb/ /app/internal/services/room_service/pb/


WORKDIR /app/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./cmd/main.go


FROM alpine:latest


RUN mkdir -p /app/config


COPY --from=builder /api-gateway /app/

COPY --from=builder /app/api/config/config.yaml /app/config/


WORKDIR /app


EXPOSE 8080 8081


CMD ["./api-gateway", "--config", "/app/config/config.yaml"]