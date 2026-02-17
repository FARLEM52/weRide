package room_service_transport

import (
	"context"
	"we_ride/internal/pkg/logger"
	roomservice "we_ride/internal/services/room_service/pb"

	"google.golang.org/grpc"
)

type Service interface {
}

type Server struct {
	roomservice.UnimplementedRoomServiceServer
	//service Service
}

func New() *Server {
	//return &Server{service: s}
	return &Server{}
}

func RegisterServer(gRPC *grpc.Server) {
	roomservice.RegisterRoomServiceServer(gRPC, &Server{})
}

func (s Server) CreateRoom(ctx context.Context,
	CreateRoomRequest *roomservice.CreateRoomRequest,
) (*roomservice.CreateRoomResponse, error) {
	logger.GetLoggerFromCtx(ctx).Info(ctx, "No Implementation")
	return nil, nil
}

func (s Server) JoinRoom(ctx context.Context,
	CreateRoomRequest *roomservice.JoinRoomRequest,
) (*roomservice.JoinRoomResponse, error) {
	logger.GetLoggerFromCtx(ctx).Info(ctx, "No Implementation")
	return nil, nil
}

func (s Server) ExitRoom(ctx context.Context,
	CreateRoomRequest *roomservice.ExitRoomRequest,
) (*roomservice.ExitRoomResponse, error) {
	logger.GetLoggerFromCtx(ctx).Info(ctx, "No Implementation")
	return nil, nil
}

func (s Server) FindRoom(ctx context.Context,
	CreateRoomRequest *roomservice.FindRoomRequest,
) (*roomservice.FindRoomResponse, error) {
	logger.GetLoggerFromCtx(ctx).Info(ctx, "No Implementation")
	return nil, nil
}

func (s Server) GetRoomDetails(ctx context.Context,
	CreateRoomRequest *roomservice.GetRoomDetailsRequest,
) (*roomservice.GetRoomDetailsResponse, error) {
	logger.GetLoggerFromCtx(ctx).Info(ctx, "No Implementation")
	return nil, nil
}

func (s Server) StreamRoomUpdates(*roomservice.StreamRoomUpdatesRequest,
	grpc.ServerStreamingServer[roomservice.RoomUpdate],
) error {
	return nil
}
