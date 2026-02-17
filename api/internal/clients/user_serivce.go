package clients

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "we_ride/internal/services/user_service/protoc/gen/go"
)

type UserServiceClient struct {
	client pb.AuthClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient(addr string) (*UserServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("fail to dial: %v", err)
	}
	client := pb.NewAuthClient(conn)
	return &UserServiceClient{client, conn}, nil
}

func (u *UserServiceClient) Login(ctx context.Context, email string, password string) (string, error) {
	resp, err := u.client.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("fail to login: %v", err)
	}
	return resp.GetToken(), nil
}

func (u *UserServiceClient) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return u.client.Register(ctx, req)
}

func (u *UserServiceClient) HistoryOfRoutes(ctx context.Context) (*pb.HistoryOfRoutesResponse, error) {
	return u.client.HistoryOfRoutes(ctx, &pb.HistoryOfRoutesRequest{})
}

func (u *UserServiceClient) Close() {
	if u.conn != nil {
		_ = u.conn.Close()
	}
}
