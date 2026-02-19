package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/metadata"

	"we_ride/api/internal/clients"
	pb_payment "we_ride/internal/services/payment_service/pb"
	pb_room "we_ride/internal/services/room_service/pb"
	pb "we_ride/internal/services/user_service/protoc/gen/go"
)

type APIHandler struct {
	userService    *clients.UserServiceClient
	roomService    *clients.RoomServiceClient
	paymentService *clients.PaymentServiceClient
}

func NewAPIHandler(
	userService *clients.UserServiceClient,
	roomService *clients.RoomServiceClient,
	paymentService *clients.PaymentServiceClient,
) *APIHandler {
	return &APIHandler{
		userService:    userService,
		roomService:    roomService,
		paymentService: paymentService,
	}
}

// getUserIDFromCtx извлекает user_id из JWT-клеймов контекста echo
func getUserIDFromCtx(c echo.Context) (string, error) {
	claims, ok := c.Get("user").(jwt.MapClaims)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
	}
	userID, ok := claims["uid"].(string)
	if !ok || userID == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}
	return userID, nil
}

// ===== Auth =====

func (h *APIHandler) Register(c echo.Context) error {
	var req pb.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	resp, err := h.userService.Register(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to register"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) Login(c echo.Context) error {
	var req pb.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	token, err := h.userService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (h *APIHandler) HistoryOfRoutes(c echo.Context) error {
	userID, err := getUserIDFromCtx(c)
	if err != nil {
		return err
	}
	ctx := metadata.AppendToOutgoingContext(c.Request().Context(), "user_id", userID)
	resp, err := h.userService.HistoryOfRoutes(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get routes"})
	}
	return c.JSON(http.StatusOK, resp)
}

// ===== Rooms =====

func (h *APIHandler) CreateRoom(c echo.Context) error {
	userID, err := getUserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req pb_room.CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	req.CreatorId = userID
	resp, err := h.roomService.CreateRoom(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create room"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) JoinRoom(c echo.Context) error {
	userID, err := getUserIDFromCtx(c)
	if err != nil {
		return err
	}
	roomID := c.Param("id")
	if roomID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Room ID is required"})
	}
	resp, err := h.roomService.JoinRoom(c.Request().Context(), &pb_room.JoinRoomRequest{RoomId: roomID, UserId: userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to join room"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) ExitRoom(c echo.Context) error {
	userID, err := getUserIDFromCtx(c)
	if err != nil {
		return err
	}
	roomID := c.Param("id")
	if roomID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Room ID is required"})
	}
	resp, err := h.roomService.ExitRoom(c.Request().Context(), &pb_room.ExitRoomRequest{RoomId: roomID, UserId: userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to exit room"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) FindRoom(c echo.Context) error {
	resp, err := h.roomService.FindRoom(c.Request().Context(), &pb_room.FindRoomRequest{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find rooms"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) GetRoomDetails(c echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Room ID is required"})
	}
	resp, err := h.roomService.GetRoomDetails(c.Request().Context(), &pb_room.GetRoomDetailsRequest{RoomId: roomID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get room details"})
	}
	return c.JSON(http.StatusOK, resp)
}

// ===== Payments =====

// ProcessPayment — POST /payments/process
// Вызывается после завершения поездки (когда room статус = COMPLETED)
// Body: { "room_id": "...", "user_ids": ["..."], "amount_per_user": 500.00, "description": "..." }
func (h *APIHandler) ProcessPayment(c echo.Context) error {
	var req pb_payment.ProcessPaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if req.RoomId == "" || len(req.UserIds) == 0 || req.AmountPerUser <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "room_id, user_ids and amount_per_user are required"})
	}
	resp, err := h.paymentService.ProcessPayment(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process payment"})
	}
	return c.JSON(http.StatusOK, resp)
}

// RefundPayment — POST /payments/refund
// Вызывается при отмене поездки
// Body: { "room_id": "...", "reason": "..." }
func (h *APIHandler) RefundPayment(c echo.Context) error {
	var req pb_payment.RefundPaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if req.RoomId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "room_id is required"})
	}
	resp, err := h.paymentService.RefundPayment(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to refund payment"})
	}
	return c.JSON(http.StatusOK, resp)
}

// GetPaymentHistory — GET /payments/history
// Возвращает историю транзакций текущего пользователя
func (h *APIHandler) GetPaymentHistory(c echo.Context) error {
	userID, err := getUserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.paymentService.GetPaymentHistory(c.Request().Context(), &pb_payment.GetPaymentHistoryRequest{UserId: userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get payment history"})
	}
	return c.JSON(http.StatusOK, resp)
}

// CompleteRide — POST /rooms/:id/complete
// Вызывается водителем после завершения поездки.
// Триггерит сохранение маршрута и автоматическую оплату.
// Body: { "driver_id": "...", "total_price": 1200.0, "distance_km": 15.5 }
func (h *APIHandler) CompleteRide(c echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "room_id is required"})
	}
	var body struct {
		TotalPrice float32 `json:"total_price"`
		DistanceKm float32 `json:"distance_km"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	driverID, err := getUserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.roomService.CompleteRide(c.Request().Context(), &pb_room.CompleteRideRequest{
		RoomId:     roomID,
		DriverId:   driverID,
		TotalPrice: body.TotalPrice,
		DistanceKm: body.DistanceKm,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to complete ride"})
	}
	return c.JSON(http.StatusOK, resp)
}
