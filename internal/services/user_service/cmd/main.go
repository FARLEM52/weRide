package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"

	"we_ride/internal/services/user_service/db/postgres"
	"we_ride/internal/services/user_service/internal/config"
	"we_ride/internal/services/user_service/internal/repository"
	"we_ride/internal/services/user_service/internal/service"
	pb "we_ride/internal/services/user_service/protoc/gen/go"

	"we_ride/internal/pkg/logger"
)

func main() {

	ctx := context.Background()

	ctx, err := logger.New(ctx)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	cfg, err := config.New()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "failed to init config: %v", zap.Error(err))
	}

	pool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "failed to init repository: %v", zap.Error(err))
	}
	fmt.Printf("DEBUG TTL value: '%s'\n", cfg.JWTAccessTokenTTL)
	ttlToken := cfg.JWTAccessTokenTTL
	repo := repository.NewRepository(pool, ttlToken, cfg.JwtSecret)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor(ctx, logger.GetLoggerFromCtx(ctx))))

	srv := service.New(repo)
	pb.RegisterAuthServer(grpcServer, srv)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", cfg.GRPCPort))
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed to listen", zap.Error(err))
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx, "gRPC server listening on "+cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
