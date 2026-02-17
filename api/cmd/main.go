package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"we_ride/api/internal/clients"
	"we_ride/api/internal/config"
	"we_ride/api/internal/router"
	"we_ride/internal/pkg/logger"
)

func main() {
	ctx, err := logger.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.New()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "failed to load config", zap.Error(err))
	}

	userClient, err := clients.NewUserServiceClient(cfg.UserServiceAddr)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "failed to create gRPC client", zap.Error(err))
	}
	defer userClient.Close()

	e := echo.New()
	
	router.InitRoutes(e, userClient, []byte(cfg.JWTSecret))

	go func() {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "starting HTTP server", zap.String("addr", cfg.RestPort))
		if err := e.Start(":" + cfg.RestPort); err != nil {
			logger.GetLoggerFromCtx(ctx).Fatal(ctx, "failed to start HTTP server", zap.Error(err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.GetLoggerFromCtx(ctx).Info(ctx, "shutting down server...")

	if err := e.Shutdown(ctx); err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "server forced to shutdown", zap.Error(err))
	}
}
