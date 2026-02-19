package service

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"we_ride/internal/services/user_service/internal/repository"
	pb "we_ride/internal/services/user_service/protoc/gen/go"
)

type ServerAPI struct {
	pb.UnimplementedAuthServer
	repo repository.Repository
}

func New(repo repository.Repository) *ServerAPI {
	return &ServerAPI{repo: repo}
}

func (s *ServerAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}
	token, err := s.repo.LoginUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.LoginResponse{Token: token}, nil
}

func (s *ServerAPI) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}
	if req.GetFirstName() == "" {
		return nil, status.Error(codes.InvalidArgument, "First name is required")
	}
	if req.GetLastName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Last name is required")
	}
	if req.GetGender() == 0 {
		return nil, status.Error(codes.InvalidArgument, "Gender is required")
	}
	userID, err := s.repo.SaveUser(ctx, req.GetEmail(), req.GetPassword(), req.GetFirstName(), req.GetLastName(), req.GetGender())
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return &pb.RegisterResponse{UserId: userID}, nil
}

func (s *ServerAPI) HistoryOfRoutes(ctx context.Context, req *pb.HistoryOfRoutesRequest) (*pb.HistoryOfRoutesResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is missing")
	}
	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in metadata")
	}
	userID, err := uuid.Parse(userIDs[0])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	routes, err := s.repo.GetUserRoutes(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get routes: %v", err)
	}

	resp := &pb.HistoryOfRoutesResponse{}
	for _, route := range routes {
		resp.Routes = append(resp.Routes, &pb.Route{
			RouteId:    route.GetRouteId(),
			DriverId:   route.GetDriverId(),
			TotalPrice: route.GetTotalPrice(),
			StartPoint: route.GetStartPoint(),
			EndPoint:   route.GetEndPoint(),
			Distance:   route.GetDistance(),
		})
	}
	return resp, nil
}

// SaveRoute — вызывается room_service при завершении поездки (COMPLETED)
func (s *ServerAPI) SaveRoute(ctx context.Context, req *pb.SaveRouteRequest) (*pb.SaveRouteResponse, error) {
	if req.GetRoomId() == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id is required")
	}
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	if len(req.GetPassengerIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "passenger_ids must not be empty")
	}

	routeID, err := s.repo.SaveRoute(ctx,
		req.GetRoomId(),
		req.GetDriverId(),
		req.GetStartPoint(),
		req.GetEndPoint(),
		req.GetDistance(),
		req.GetTotalPrice(),
		req.GetPassengerIds(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save route: %v", err)
	}
	return &pb.SaveRouteResponse{RouteId: routeID}, nil
}
