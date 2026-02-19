package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"we_ride/internal/services/payment_service/internal/repository"
	"we_ride/internal/services/payment_service/internal/yookassa"
	pb "we_ride/internal/services/payment_service/pb"
)

const currency = "RUB"

type yookassaClient interface {
	CreatePayment(ctx context.Context, idempotencyKey string, req yookassa.CreatePaymentRequest) (*yookassa.PaymentResponse, error)
	CreateRefund(ctx context.Context, idempotencyKey string, req yookassa.CreateRefundRequest) (*yookassa.RefundResponse, error)
}

type PaymentService struct {
	pb.UnimplementedPaymentServiceServer
	repo     repository.Repository
	yookassa yookassaClient
}

func New(repo repository.Repository, ykClient yookassaClient) *PaymentService {
	return &PaymentService{repo: repo, yookassa: ykClient}
}

// ProcessPayment — создаёт платёж в ЮKassa для каждого пассажира комнаты
func (s *PaymentService) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	if req.RoomId == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id is required")
	}
	if len(req.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_ids must not be empty")
	}
	if req.AmountPerUser <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount_per_user must be greater than 0")
	}

	var results []*pb.Payment
	amountStr := fmt.Sprintf("%.2f", req.AmountPerUser)

	for _, userID := range req.UserIds {
		paymentID := uuid.New().String()
		idempotencyKey := fmt.Sprintf("%s-%s", req.RoomId, userID)

		description := req.Description
		if description == "" {
			description = fmt.Sprintf("Оплата поездки (комната %s)", req.RoomId)
		}

		// Создаём платёж в ЮKassa (или мок-режим, если клиент не настроен)
		paymentStatus := "succeeded"
		yookassaID := ""
		var err error
		if s.yookassa != nil {
			ykResp, ykErr := s.yookassa.CreatePayment(ctx, idempotencyKey, yookassa.CreatePaymentRequest{
				Amount: yookassa.Amount{
					Value:    amountStr,
					Currency: currency,
				},
				Confirmation: yookassa.Confirmation{
					Type:      "redirect",
					ReturnURL: "https://weride.app/payment/success",
				},
				Description: description,
				Capture:     true,
				Metadata: map[string]string{
					"room_id":    req.RoomId,
					"user_id":    userID,
					"payment_id": paymentID,
				},
			})
			err = ykErr
			if ykErr != nil {
				paymentStatus = "failed"
			} else {
				yookassaID = ykResp.ID
				paymentStatus = ykResp.Status
			}
		} else {
			yookassaID = "mock-" + paymentID
		}

		// Сохраняем в БД в любом случае (даже при ошибке — для аудита)
		record := &repository.PaymentRecord{
			PaymentID:         paymentID,
			RoomID:            req.RoomId,
			UserID:            userID,
			Amount:            float64(req.AmountPerUser),
			Currency:          currency,
			Status:            paymentStatus,
			YookassaPaymentID: yookassaID,
			Description:       description,
		}

		if saveErr := s.repo.CreatePayment(ctx, record); saveErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to save payment: %v", saveErr)
		}

		if err != nil {
			return nil, status.Errorf(codes.Internal, "yookassa error for user %s: %v", userID, err)
		}

		results = append(results, &pb.Payment{
			PaymentId:         paymentID,
			RoomId:            req.RoomId,
			UserId:            userID,
			Amount:            req.AmountPerUser,
			Currency:          currency,
			Status:            paymentStatus,
			YookassaPaymentId: yookassaID,
			CreatedAt:         time.Now().Format(time.RFC3339),
			Description:       description,
		})
	}

	return &pb.ProcessPaymentResponse{
		Payments: results,
		Success:  true,
	}, nil
}

// RefundPayment — возвращает деньги всем пассажирам комнаты
func (s *PaymentService) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	if req.RoomId == "" {
		return nil, status.Error(codes.InvalidArgument, "room_id is required")
	}

	// Получаем все успешные платежи по комнате
	payments, err := s.repo.GetPaymentsByRoom(ctx, req.RoomId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get payments: %v", err)
	}

	var refunds []*pb.Payment

	for _, p := range payments {
		// Возвращаем только успешные платежи
		if p.Status != "succeeded" {
			continue
		}

		idempotencyKey := fmt.Sprintf("refund-%s", p.PaymentID)
		amountStr := fmt.Sprintf("%.2f", p.Amount)
		reason := req.Reason
		if reason == "" {
			reason = "Отмена поездки"
		}

		var ykErr error
		if s.yookassa != nil {
			_, ykErr = s.yookassa.CreateRefund(ctx, idempotencyKey, yookassa.CreateRefundRequest{
				PaymentID: p.YookassaPaymentID,
				Amount: yookassa.Amount{
					Value:    amountStr,
					Currency: p.Currency,
				},
				Description: reason,
			})
		}

		newStatus := "refunded"
		if ykErr != nil {
			newStatus = "refund_failed"
		}

		if updateErr := s.repo.UpdatePaymentStatus(ctx, p.PaymentID, newStatus, p.YookassaPaymentID); updateErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to update payment status: %v", updateErr)
		}

		if ykErr != nil {
			return nil, status.Errorf(codes.Internal, "yookassa refund error for payment %s: %v", p.PaymentID, ykErr)
		}

		refunds = append(refunds, &pb.Payment{
			PaymentId:         p.PaymentID,
			RoomId:            p.RoomID,
			UserId:            p.UserID,
			Amount:            float32(p.Amount),
			Currency:          p.Currency,
			Status:            newStatus,
			YookassaPaymentId: p.YookassaPaymentID,
			CreatedAt:         p.CreatedAt.Format(time.RFC3339),
			Description:       reason,
		})
	}

	return &pb.RefundPaymentResponse{
		Refunds: refunds,
		Success: true,
	}, nil
}

// GetPaymentHistory — история транзакций пользователя
func (s *PaymentService) GetPaymentHistory(ctx context.Context, req *pb.GetPaymentHistoryRequest) (*pb.GetPaymentHistoryResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	records, err := s.repo.GetPaymentsByUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get payment history: %v", err)
	}

	var payments []*pb.Payment
	for _, p := range records {
		payments = append(payments, &pb.Payment{
			PaymentId:         p.PaymentID,
			RoomId:            p.RoomID,
			UserId:            p.UserID,
			Amount:            float32(p.Amount),
			Currency:          p.Currency,
			Status:            p.Status,
			YookassaPaymentId: p.YookassaPaymentID,
			CreatedAt:         p.CreatedAt.Format(time.RFC3339),
			Description:       p.Description,
		})
	}

	return &pb.GetPaymentHistoryResponse{Payments: payments}, nil
}
