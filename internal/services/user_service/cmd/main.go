package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"we_ride/internal/services/user_service/db/postgres"
	"we_ride/internal/services/user_service/internal/config"
	"we_ride/internal/services/user_service/internal/repository"
	"we_ride/internal/services/user_service/internal/service"
	pb "we_ride/internal/services/user_service/protoc/gen/go"

	"we_ride/internal/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, err := logger.New(ctx)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	log := logger.GetLoggerFromCtx(ctx)

	cfg, err := config.New()
	if err != nil {
		log.Fatal(ctx, "failed to init config", zap.Error(err))
	}

	pool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		log.Fatal(ctx, "failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	repo := repository.NewRepository(pool, cfg.JWTAccessTokenTTL, cfg.JwtSecret)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor(ctx, log)))

	srv := service.New(repo)
	pb.RegisterAuthServer(grpcServer, srv)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", cfg.GRPCPort))
	if err != nil {
		log.Fatal(ctx, "failed to listen", zap.Error(err))
	}

	log.Info(ctx, "user service gRPC server started", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(ctx, "failed to serve gRPC", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info(ctx, "shutting down gracefully...")
	grpcServer.GracefulStop()
	log.Info(ctx, "user service stopped")
}
