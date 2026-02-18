package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PaymentRecord — модель платежа в БД
type PaymentRecord struct {
	PaymentID         string
	RoomID            string
	UserID            string
	Amount            float64
	Currency          string
	Status            string
	YookassaPaymentID string
	Description       string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Repository interface {
	CreatePayment(ctx context.Context, p *PaymentRecord) error
	UpdatePaymentStatus(ctx context.Context, paymentID, status, yookassaID string) error
	GetPaymentsByRoom(ctx context.Context, roomID string) ([]*PaymentRecord, error)
	GetPaymentsByUser(ctx context.Context, userID string) ([]*PaymentRecord, error)
	GetPaymentByID(ctx context.Context, paymentID string) (*PaymentRecord, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) CreatePayment(ctx context.Context, p *PaymentRecord) error {
	query := `
		INSERT INTO payments (payment_id, room_id, user_id, amount, currency, status, yookassa_payment_id, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		p.PaymentID, p.RoomID, p.UserID,
		p.Amount, p.Currency, p.Status,
		p.YookassaPaymentID, p.Description,
	)
	if err != nil {
		return fmt.Errorf("CreatePayment: %w", err)
	}
	return nil
}

func (r *repository) UpdatePaymentStatus(ctx context.Context, paymentID, status, yookassaID string) error {
	query := `
		UPDATE payments
		SET status = $1, yookassa_payment_id = $2, updated_at = NOW()
		WHERE payment_id = $3
	`
	_, err := r.db.Exec(ctx, query, status, yookassaID, paymentID)
	if err != nil {
		return fmt.Errorf("UpdatePaymentStatus: %w", err)
	}
	return nil
}

func (r *repository) GetPaymentsByRoom(ctx context.Context, roomID string) ([]*PaymentRecord, error) {
	query := `
		SELECT payment_id, room_id, user_id, amount, currency, status,
		       yookassa_payment_id, description, created_at, updated_at
		FROM payments WHERE room_id = $1
		ORDER BY created_at DESC
	`
	return r.scanPayments(ctx, query, roomID)
}

func (r *repository) GetPaymentsByUser(ctx context.Context, userID string) ([]*PaymentRecord, error) {
	query := `
		SELECT payment_id, room_id, user_id, amount, currency, status,
		       yookassa_payment_id, description, created_at, updated_at
		FROM payments WHERE user_id = $1
		ORDER BY created_at DESC
	`
	return r.scanPayments(ctx, query, userID)
}

func (r *repository) GetPaymentByID(ctx context.Context, paymentID string) (*PaymentRecord, error) {
	query := `
		SELECT payment_id, room_id, user_id, amount, currency, status,
		       yookassa_payment_id, description, created_at, updated_at
		FROM payments WHERE payment_id = $1
	`
	row := r.db.QueryRow(ctx, query, paymentID)
	p := &PaymentRecord{}
	err := row.Scan(
		&p.PaymentID, &p.RoomID, &p.UserID,
		&p.Amount, &p.Currency, &p.Status,
		&p.YookassaPaymentID, &p.Description,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetPaymentByID: %w", err)
	}
	return p, nil
}

func (r *repository) scanPayments(ctx context.Context, query string, arg interface{}) ([]*PaymentRecord, error) {
	rows, err := r.db.Query(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var result []*PaymentRecord
	for rows.Next() {
		p := &PaymentRecord{}
		err := rows.Scan(
			&p.PaymentID, &p.RoomID, &p.UserID,
			&p.Amount, &p.Currency, &p.Status,
			&p.YookassaPaymentID, &p.Description,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}
