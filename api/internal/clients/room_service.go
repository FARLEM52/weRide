package clients

import (
	"context"
	"fmt"

	pb "we_ride/internal/services/room_service/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RoomServiceClient struct {
	client pb.RoomServiceClient
	conn   *grpc.ClientConn
}

func NewRoomServiceClient(addr string) (*RoomServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("fail to create grpc client for room service: %v", err)
	}
	return &RoomServiceClient{client: pb.NewRoomServiceClient(conn), conn: conn}, nil
}

func (r *RoomServiceClient) Close() {
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func (r *RoomServiceClient) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	resp, err := r.client.CreateRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CreateRoom: %w", err)
	}
	return resp, nil
}

func (r *RoomServiceClient) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.JoinRoomResponse, error) {
	resp, err := r.client.JoinRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("JoinRoom: %w", err)
	}
	return resp, nil
}

func (r *RoomServiceClient) ExitRoom(ctx context.Context, req *pb.ExitRoomRequest) (*pb.ExitRoomResponse, error) {
	resp, err := r.client.ExitRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ExitRoom: %w", err)
	}
	return resp, nil
}

func (r *RoomServiceClient) FindRoom(ctx context.Context, req *pb.FindRoomRequest) (*pb.FindRoomResponse, error) {
	resp, err := r.client.FindRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("FindRoom: %w", err)
	}
	return resp, nil
}

func (r *RoomServiceClient) GetRoomDetails(ctx context.Context, req *pb.GetRoomDetailsRequest) (*pb.GetRoomDetailsResponse, error) {
	resp, err := r.client.GetRoomDetails(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetRoomDetails: %w", err)
	}
	return resp, nil
}

// CompleteRide вызывает метод через прямой Invoke, т.к. он не входит в сгенерированный pb клиент
func (r *RoomServiceClient) CompleteRide(ctx context.Context, req *pb.CompleteRideRequest) (*pb.CompleteRideResponse, error) {
	resp, err := pb.CompleteRideClient(r.conn, ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CompleteRide: %w", err)
	}
	return resp, nil
}
