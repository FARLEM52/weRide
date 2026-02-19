package clients

import (
	"context"
	"fmt"

	pb "we_ride/internal/services/payment_service/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentServiceClient struct {
	client pb.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewPaymentServiceClient(addr string) (*PaymentServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("fail to create grpc client for payment service: %v", err)
	}
	return &PaymentServiceClient{client: pb.NewPaymentServiceClient(conn), conn: conn}, nil
}

func (p *PaymentServiceClient) Close() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
}

func (p *PaymentServiceClient) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	resp, err := p.client.ProcessPayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ProcessPayment: %w", err)
	}
	return resp, nil
}

func (p *PaymentServiceClient) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	resp, err := p.client.RefundPayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("RefundPayment: %w", err)
	}
	return resp, nil
}

func (p *PaymentServiceClient) GetPaymentHistory(ctx context.Context, req *pb.GetPaymentHistoryRequest) (*pb.GetPaymentHistoryResponse, error) {
	resp, err := p.client.GetPaymentHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetPaymentHistory: %w", err)
	}
	return resp, nil
}
