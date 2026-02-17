package service

import (
	"context"
	"fmt"
	"we_ride/internal/services/room_service/internal/repository"
	roomservice "we_ride/internal/services/room_service/pb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RoomService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) CreateRoom(ctx context.Context, req *roomservice.CreateRoomRequest) (*roomservice.CreateRoomResponse, error) {
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
		return nil, err
	}

	// создатель сразу добавляется в участники
	if err := s.repo.AddMember(ctx, roomID, req.CreatorId); err != nil {
		return nil, err
	}

	return &roomservice.CreateRoomResponse{Room: room}, nil
}

func (s *RoomService) JoinRoom(ctx context.Context, req *roomservice.JoinRoomRequest) (*roomservice.JoinRoomResponse, error) {
	room, err := s.repo.GetRoomByID(ctx, req.RoomId)
	if err != nil {
		return nil, err
	}

	if len(room.Members) >= int(room.AvailableSeats) {
		room.Status = roomservice.RoomStatus_ROOM_STATUS_FULL
		s.repo.UpdateRoomStatus(ctx, room.RoomId, room.Status)
		return nil, fmt.Errorf("room is full")
	}

	if err := s.repo.AddMember(ctx, req.RoomId, req.UserId); err != nil {
		return nil, err
	}

	return &roomservice.JoinRoomResponse{Room: room}, nil
}

func (s *RoomService) ExitRoom(ctx context.Context, req *roomservice.ExitRoomRequest) (*roomservice.ExitRoomResponse, error) {
	if err := s.repo.RemoveMember(ctx, req.RoomId, req.UserId); err != nil {
		return nil, err
	}
	return &roomservice.ExitRoomResponse{Success: true}, nil
}

func (s *RoomService) FindRoom(ctx context.Context, req *roomservice.FindRoomRequest) (*roomservice.FindRoomResponse, error) {
	rooms, err := s.repo.ListAvailableRooms(ctx)
	if err != nil {
		return nil, err
	}
	return &roomservice.FindRoomResponse{AvailableRooms: rooms}, nil
}

func (s *RoomService) GetRoomDetails(ctx context.Context, req *roomservice.GetRoomDetailsRequest) (*roomservice.GetRoomDetailsResponse, error) {
	room, err := s.repo.GetRoomByID(ctx, req.RoomId)
	if err != nil {
		return nil, err
	}
	return &roomservice.GetRoomDetailsResponse{Room: room}, nil
}
