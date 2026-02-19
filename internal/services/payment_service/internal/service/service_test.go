package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"we_ride/internal/services/payment_service/internal/repository"
	"we_ride/internal/services/payment_service/internal/yookassa"
	pb "we_ride/internal/services/payment_service/pb"
)

type fakePaymentRepo struct {
	created []*repository.PaymentRecord
	byRoom  []*repository.PaymentRecord
	byUser  []*repository.PaymentRecord
	updated map[string]string
}

func (f *fakePaymentRepo) CreatePayment(_ context.Context, p *repository.PaymentRecord) error {
	f.created = append(f.created, p)
	return nil
}
func (f *fakePaymentRepo) UpdatePaymentStatus(_ context.Context, paymentID, status, _ string) error {
	if f.updated == nil {
		f.updated = map[string]string{}
	}
	f.updated[paymentID] = status
	return nil
}
func (f *fakePaymentRepo) GetPaymentsByRoom(_ context.Context, _ string) ([]*repository.PaymentRecord, error) {
	return f.byRoom, nil
}
func (f *fakePaymentRepo) GetPaymentsByUser(_ context.Context, _ string) ([]*repository.PaymentRecord, error) {
	return f.byUser, nil
}
func (f *fakePaymentRepo) GetPaymentByID(_ context.Context, _ string) (*repository.PaymentRecord, error) {
	return nil, errors.New("not implemented")
}

type fakeYookassa struct {
	createErr error
	refundErr error
}

func (f *fakeYookassa) CreatePayment(_ context.Context, _ string, _ yookassa.CreatePaymentRequest) (*yookassa.PaymentResponse, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &yookassa.PaymentResponse{ID: "yk-payment-id", Status: "succeeded"}, nil
}

func (f *fakeYookassa) CreateRefund(_ context.Context, _ string, _ yookassa.CreateRefundRequest) (*yookassa.RefundResponse, error) {
	if f.refundErr != nil {
		return nil, f.refundErr
	}
	return &yookassa.RefundResponse{ID: "yk-refund-id", Status: "succeeded"}, nil
}

func TestProcessPaymentValidation(t *testing.T) {
	svc := New(&fakePaymentRepo{}, nil)

	_, err := svc.ProcessPayment(context.Background(), &pb.ProcessPaymentRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestProcessPaymentMockSuccess(t *testing.T) {
	repo := &fakePaymentRepo{}
	svc := New(repo, nil)

	resp, err := svc.ProcessPayment(context.Background(), &pb.ProcessPaymentRequest{
		RoomId:        "room-1",
		UserIds:       []string{"u1", "u2"},
		AmountPerUser: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success || len(resp.Payments) != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if len(repo.created) != 2 {
		t.Fatalf("expected 2 persisted payments, got %d", len(repo.created))
	}
	if repo.created[0].Status != "succeeded" {
		t.Fatalf("expected succeeded status in mock mode, got %s", repo.created[0].Status)
	}
}

func TestProcessPaymentYookassaErrorPersistsAudit(t *testing.T) {
	repo := &fakePaymentRepo{}
	svc := New(repo, &fakeYookassa{createErr: errors.New("gateway down")})

	_, err := svc.ProcessPayment(context.Background(), &pb.ProcessPaymentRequest{
		RoomId:        "room-1",
		UserIds:       []string{"u1"},
		AmountPerUser: 10,
	})
	if err == nil {
		t.Fatal("expected yookassa error")
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected one persisted failed payment, got %d", len(repo.created))
	}
	if repo.created[0].Status != "failed" {
		t.Fatalf("expected failed status, got %s", repo.created[0].Status)
	}
}

func TestRefundPaymentScenarios(t *testing.T) {
	repo := &fakePaymentRepo{byRoom: []*repository.PaymentRecord{
		{PaymentID: "p1", RoomID: "room-1", UserID: "u1", Amount: 100, Currency: "RUB", Status: "succeeded", YookassaPaymentID: "yk1", CreatedAt: time.Now()},
		{PaymentID: "p2", RoomID: "room-1", UserID: "u2", Amount: 100, Currency: "RUB", Status: "failed", YookassaPaymentID: "yk2", CreatedAt: time.Now()},
	}}
	svc := New(repo, &fakeYookassa{})

	resp, err := svc.RefundPayment(context.Background(), &pb.RefundPaymentRequest{RoomId: "room-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success || len(resp.Refunds) != 1 {
		t.Fatalf("unexpected refund response: %+v", resp)
	}
	if repo.updated["p1"] != "refunded" {
		t.Fatalf("expected p1 to be refunded, got %s", repo.updated["p1"])
	}
}

func TestGetPaymentHistory(t *testing.T) {
	repo := &fakePaymentRepo{byUser: []*repository.PaymentRecord{
		{PaymentID: "p1", RoomID: "room-1", UserID: "u1", Amount: 100, Currency: "RUB", Status: "succeeded", Description: "trip", CreatedAt: time.Now()},
	}}
	svc := New(repo, nil)

	resp, err := svc.GetPaymentHistory(context.Background(), &pb.GetPaymentHistoryRequest{UserId: "u1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Payments) != 1 {
		t.Fatalf("expected 1 payment in history, got %d", len(resp.Payments))
	}
}
