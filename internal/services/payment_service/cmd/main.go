package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"we_ride/internal/pkg/logger"
	"we_ride/internal/services/payment_service/config"
	"we_ride/internal/services/payment_service/database"
	"we_ride/internal/services/payment_service/internal/repository"
	"we_ride/internal/services/payment_service/internal/service"
	"we_ride/internal/services/payment_service/internal/yookassa"
	pb "we_ride/internal/services/payment_service/pb"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, err := logger.New(ctx)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	l := logger.GetLoggerFromCtx(ctx)

	cfg, err := config.New()
	if err != nil {
		l.Fatal(ctx, "failed to init config", zap.Error(err))
	}

	if cfg.YookassaShopID == "" || cfg.YookassaSecretKey == "" {
		l.Fatal(ctx, "YOOKASSA_SHOP_ID and YOOKASSA_SECRET_KEY must be set")
	}

	if err := database.RunMigrations(ctx, cfg.Postgres); err != nil {
		l.Fatal(ctx, "failed to run migrations", zap.Error(err))
	}

	pool, err := database.New(ctx, cfg.Postgres)
	if err != nil {
		l.Fatal(ctx, "failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	ykClient := yookassa.NewClient(cfg.YookassaShopID, cfg.YookassaSecretKey)
	repo := repository.NewRepository(pool)
	svc := service.New(repo, ykClient)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor(ctx, l)))
	pb.RegisterPaymentServiceServer(grpcServer, svc)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.GRPCHost, cfg.GRPCPort))
	if err != nil {
		l.Fatal(ctx, "failed to listen", zap.Error(err))
	}

	l.Info(ctx, "payment service gRPC server started", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal(ctx, "failed to serve gRPC", zap.Error(err))
		}
	}()

	<-ctx.Done()
	l.Info(ctx, "shutting down gracefully...")
	grpcServer.GracefulStop()
	l.Info(ctx, "payment service stopped")
}
