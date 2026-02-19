package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type key string

const (
	Key = key("logger")

	RequestID = key("request_id")
)

type Logger struct {
	logger *zap.Logger
}

func New(ctx context.Context) (context.Context, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("logger.NewLogger: %w", err)
	}

	ctx = context.WithValue(ctx, Key, &Logger{logger: logger})

	return ctx, nil
}

func GetLoggerFromCtx(ctx context.Context) *Logger {
	return ctx.Value(Key).(*Logger)
}

func Interceptor(ctx context.Context, logger *Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		guid := uuid.New().String()
		ctx = context.WithValue(ctx, RequestID, guid)
		ctx = context.WithValue(ctx, Key, logger)

		logger := GetLoggerFromCtx(ctx)

		logger.Info(ctx,
			"request", zap.String("method", info.FullMethod),
			zap.Time("request time", time.Now()),
		)

		return handler(ctx, req)
	}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}

	l.logger.Info(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}

	l.logger.Fatal(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}

	l.logger.Error(msg, fields...)
}
