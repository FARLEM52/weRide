package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	paymentpb "we_ride/internal/services/payment_service/pb"
	"we_ride/internal/services/room_service/internal/repository"
	roomservice "we_ride/internal/services/room_service/pb"
	authpb "we_ride/internal/services/user_service/protoc/gen/go"
)

type paymentProcessor func(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error)
type routeSaver func(req *roomservice.CompleteRideRequest, memberIDs []string, startAddr, endAddr string, totalPrice float32)

type RoomService struct {
	roomservice.UnimplementedRoomServiceServer
	repo               repository.Repository
	userServiceAddr    string
	paymentServiceAddr string

	processPaymentFn paymentProcessor
	saveRouteFn      routeSaver
}

func New(repo repository.Repository, userServiceAddr, paymentServiceAddr string) *RoomService {
	return &RoomService{
		repo:               repo,
		userServiceAddr:    userServiceAddr,
		paymentServiceAddr: paymentServiceAddr,
	}
}

func (s *RoomService) CreateRoom(ctx context.Context, req *roomservice.CreateRoomRequest) (*roomservice.CreateRoomResponse, error) {
	if req.StartLocation == nil || req.EndLocation == nil {
		return nil, status.Error(codes.InvalidArgument, "start and end location are required")
	}
	if req.CreatorId == "" {
		return nil, status.Error(codes.InvalidArgument, "creator_id is required")
	}
	if req.MaxMembers <= 0 {
		return nil, status.Error(codes.InvalidArgument, "max_members must be greater than 0")
	}

	roomID := uuid.New().String()
	room := &roomservice.Room{
		RoomId:         roomID,
		CreatorId:      req.CreatorId,
		AvailableSeats: req.MaxMembers,
		Status:         roomservice.RoomStatus_ROOM_STATUS_WAITING,
		StartLocation:  req.StartLocation,
		EndLocation:    req.EndLocation,
		CreatedAt:      timestamppb.Now(),
		ScheduledTime:  req.ScheduledTime,
		TotalPrice:     0,
		CostPerMember:  0,
	}

	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create room: %v", err)
	}
	if err := s.repo.AddMember(ctx, roomID, req.CreatorId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add creator as member: %v", err)
	}
	return &roomservice.CreateRoomResponse{Room: room}, nil
}

func (s *RoomService) JoinRoom(ctx context.Context, req *roomservice.JoinRoomRequest) (*roomservice.JoinRoomResponse, error) {
	if req.RoomId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id and user_id are required")
	}
	room, err := s.repo.GetRoomByID(ctx, req.RoomId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "room not found: %v", err)
	}

	memberIDs, err := s.repo.GetRoomMembers(ctx, req.RoomId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get room members: %v", err)
	}

	if len(memberIDs) >= int(room.AvailableSeats) {
		if err := s.repo.UpdateRoomStatus(ctx, room.RoomId, roomservice.RoomStatus_ROOM_STATUS_FULL); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update room status: %v", err)
		}
		return nil, status.Error(codes.FailedPrecondition, "room is full")
	}
	if err := s.repo.AddMember(ctx, req.RoomId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to join room: %v", err)
	}
	return &roomservice.JoinRoomResponse{Room: room}, nil
}

func (s *RoomService) ExitRoom(ctx context.Context, req *roomservice.ExitRoomRequest) (*roomservice.ExitRoomResponse, error) {
	if req.RoomId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id and user_id are required")
	}
	if err := s.repo.RemoveMember(ctx, req.RoomId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to exit room: %v", err)
	}
	return &roomservice.ExitRoomResponse{Success: true}, nil
}

func (s *RoomService) FindRoom(ctx context.Context, _ *roomservice.FindRoomRequest) (*roomservice.FindRoomResponse, error) {
	rooms, err := s.repo.ListAvailableRooms(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find rooms: %v", err)
	}
	return &roomservice.FindRoomResponse{AvailableRooms: rooms}, nil
}

func (s *RoomService) GetRoomDetails(ctx context.Context, req *roomservice.GetRoomDetailsRequest) (*roomservice.GetRoomDetailsResponse, error) {
	if req.RoomId == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id is required")
	}
	room, err := s.repo.GetRoomByID(ctx, req.RoomId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "room not found: %v", err)
	}
	return &roomservice.GetRoomDetailsResponse{Room: room}, nil
}

func (s *RoomService) CompleteRide(ctx context.Context, req *roomservice.CompleteRideRequest) (*roomservice.CompleteRideResponse, error) {
	if req.RoomId == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id is required")
	}
	if req.DriverId == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}

	room, err := s.repo.GetRoomByID(ctx, req.RoomId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "room not found: %v", err)
	}
	if room.Status == roomservice.RoomStatus_ROOM_STATUS_COMPLETED {
		return nil, status.Error(codes.AlreadyExists, "ride already completed")
	}

	memberIDs, err := s.repo.GetRoomMembers(ctx, req.RoomId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get members: %v", err)
	}
	if len(memberIDs) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "no members in room")
	}

	totalPrice := req.TotalPrice
	costPerMember := totalPrice / float32(len(memberIDs))

	if err := s.repo.CompleteRoom(ctx, req.RoomId, totalPrice, costPerMember); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to complete room: %v", err)
	}

	startAddr := ""
	if room.StartLocation != nil {
		startAddr = room.StartLocation.Address
	}
	endAddr := ""
	if room.EndLocation != nil {
		endAddr = room.EndLocation.Address
	}

	s.saveRoute(req, memberIDs, startAddr, endAddr, totalPrice)

	payResp, err := s.processPayment(ctx, &paymentpb.ProcessPaymentRequest{
		RoomId:        req.RoomId,
		UserIds:       memberIDs,
		AmountPerUser: costPerMember,
		Description:   fmt.Sprintf("Поездка %s → %s", startAddr, endAddr),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "payment failed: %v", err)
	}

	return &roomservice.CompleteRideResponse{
		Success:       true,
		TotalPrice:    totalPrice,
		CostPerMember: costPerMember,
		PaymentsCount: int32(len(payResp.Payments)),
	}, nil
}

func (s *RoomService) processPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	if s.processPaymentFn != nil {
		return s.processPaymentFn(ctx, req)
	}

	paymentConn, err := grpc.NewClient(s.paymentServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}
	defer paymentConn.Close()

	paymentClient := paymentpb.NewPaymentServiceClient(paymentConn)
	return paymentClient.ProcessPayment(ctx, req)
}

func (s *RoomService) saveRoute(req *roomservice.CompleteRideRequest, memberIDs []string, startAddr, endAddr string, totalPrice float32) {
	if s.saveRouteFn != nil {
		s.saveRouteFn(req, memberIDs, startAddr, endAddr, totalPrice)
		return
	}

	go func() {
		userConn, err := grpc.NewClient(s.userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		defer userConn.Close()

		authClient := authpb.NewAuthClient(userConn)
		_, _ = authClient.SaveRoute(context.Background(), &authpb.SaveRouteRequest{
			RoomId:       req.RoomId,
			DriverId:     req.DriverId,
			StartPoint:   startAddr,
			EndPoint:     endAddr,
			Distance:     float64(req.DistanceKm),
			TotalPrice:   float64(totalPrice),
			PassengerIds: memberIDs,
		})
	}()
}

func (s *RoomService) StreamRoomUpdates(
	_ *roomservice.StreamRoomUpdatesRequest,
	_ grpc.ServerStreamingServer[roomservice.RoomUpdate],
) error {
	return status.Error(codes.Unimplemented, "StreamRoomUpdates is not implemented yet")
}
