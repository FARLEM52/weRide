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

	"we_ride/internal/pkg/logger"

	"we_ride/internal/services/room_service/config"
	"we_ride/internal/services/room_service/internal/repository"
	"we_ride/internal/services/room_service/internal/service"
	pb "we_ride/internal/services/room_service/pb"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// --- Инициализация логгера ---
	ctx, err := logger.New(ctx)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	logger := logger.GetLoggerFromCtx(ctx)

	// --- Конфиг ---
	cfg, err := config.New()
	if err != nil {
		logger.Fatal(ctx, "failed to init config", zap.Error(err))
	}

	// --- Подключение к БД ---
	pool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		logger.Fatal(ctx, "failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// --- Репозиторий и сервис ---
	repo := repository.NewRepository(pool)
	roomService := service.New(repo)

	// --- gRPC сервер ---
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor(ctx, logger)))


	pb.RegisterRoomServiceServer(grpcServer, roomService)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal(ctx, "failed to listen", zap.Error(err))
	}

	logger.Info(ctx, "Room service gRPC server started", zap.String("port", cfg.GRPCPort))

	// --- Запуск сервера в отдельной горутине ---
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal(ctx, "failed to serve gRPC", zap.Error(err))
		}
	}()

	// --- Ожидание сигнала завершения ---
	<-ctx.Done()
	logger.Info(ctx, "Shutting down gracefully...")

	grpcServer.GracefulStop()
	logger.Info(ctx, "gRPC server stopped")

	logger.Info(ctx, "Room service exited cleanly")
}
