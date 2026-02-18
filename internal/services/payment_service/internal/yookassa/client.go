package yookassa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.yookassa.ru/v2"

// PaymentStatus статусы платежа в ЮKassa
const (
	StatusPending           = "pending"
	StatusWaitingForCapture = "waiting_for_capture"
	StatusSucceeded         = "succeeded"
	StatusCanceled          = "canceled"
)

// Client HTTP-клиент для работы с ЮKassa API
type Client struct {
	shopID     string
	secretKey  string
	httpClient *http.Client
}

func NewClient(shopID, secretKey string) *Client {
	return &Client{
		shopID:     shopID,
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreatePaymentRequest тело запроса на создание платежа
type CreatePaymentRequest struct {
	Amount       Amount            `json:"amount"`
	Confirmation Confirmation      `json:"confirmation"`
	Description  string            `json:"description"`
	Capture      bool              `json:"capture"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Confirmation struct {
	Type      string `json:"type"`
	ReturnURL string `json:"return_url,omitempty"`
}

// PaymentResponse ответ от ЮKassa на создание/получение платежа
type PaymentResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Amount       Amount `json:"amount"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url,omitempty"`
	} `json:"confirmation"`
}

// CreateRefundRequest тело запроса на возврат
type CreateRefundRequest struct {
	PaymentID   string `json:"payment_id"`
	Amount      Amount `json:"amount"`
	Description string `json:"description"`
}

// RefundResponse ответ от ЮKassa на создание возврата
type RefundResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Amount    Amount `json:"amount"`
	CreatedAt string `json:"created_at"`
}

// CreatePayment создаёт новый платёж в ЮKassa
func (c *Client) CreatePayment(ctx context.Context, idempotencyKey string, req CreatePaymentRequest) (*PaymentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/payments", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.SetBasicAuth(c.shopID, c.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", idempotencyKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("yookassa error [%d]: %s", resp.StatusCode, string(respBody))
	}

	var payment PaymentResponse
	if err := json.Unmarshal(respBody, &payment); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &payment, nil
}

// GetPayment получает статус платежа по ID
func (c *Client) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/payments/"+paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.SetBasicAuth(c.shopID, c.secretKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var payment PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &payment, nil
}

// CreateRefund создаёт возврат для платежа
func (c *Client) CreateRefund(ctx context.Context, idempotencyKey string, req CreateRefundRequest) (*RefundResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/refunds", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.SetBasicAuth(c.shopID, c.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", idempotencyKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("yookassa refund error [%d]: %s", resp.StatusCode, string(respBody))
	}

	var refund RefundResponse
	if err := json.Unmarshal(respBody, &refund); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &refund, nil
}
