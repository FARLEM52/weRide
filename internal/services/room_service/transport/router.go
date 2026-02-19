package room_service_transport

import (
	"context"
	"fmt"
	"net"
	"we_ride/internal/pkg/logger"
	"we_ride/internal/services/room_service/config"
	roomservice "we_ride/internal/services/room_service/pb"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Router struct {
	Config  config.Config
	Handler Server
	Server  *grpc.Server
	Lis     *net.Listener

	RestProxyHost string
	RestProxyPort string
}

func NewRouter(ctx context.Context, cfg *config.Config, h Server) *Router {
	server := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor(ctx, logger.GetLoggerFromCtx(ctx))))
	roomservice.RegisterRoomServiceServer(server, h)

	linq := fmt.Sprintf("%s:%s", cfg.GRPCHost, cfg.GRPCPort)

	lis, err := net.Listen("tcp", linq)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed to listen", zap.Error(err))
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx, "New server config", zap.String("linq", linq))

	return &Router{
		Config: *cfg,
		Server: server,
		Lis:    &lis,

		RestProxyPort: cfg.RESTPort,
		RestProxyHost: cfg.RESTHost,
	}
}

func (r *Router) Run(ctx context.Context) {
	l := "Starting gRPC server on " + (*r.Lis).Addr().String()
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Starting gRPC", zap.String("starting", l))

	go func() {
		if err := r.Server.Serve(*r.Lis); err != nil {
			logger.GetLoggerFromCtx(ctx).Info(ctx, "failed to serve", zap.Error(err))
		}
	}()

	logger.GetLoggerFromCtx(ctx).Info(ctx, "Server is running")
}
