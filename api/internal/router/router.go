package router

import (
	"github.com/labstack/echo/v4"
	"we_ride/api/internal/clients"
	"we_ride/api/internal/handlers"
	"we_ride/api/internal/middlewares"
)

func InitRoutes(
	e *echo.Echo,
	userService *clients.UserServiceClient,
	roomService *clients.RoomServiceClient,
	paymentService *clients.PaymentServiceClient,
	jwtSecret []byte,
) {
	handler := handlers.NewAPIHandler(userService, roomService, paymentService)

	// Публичные
	e.POST("/auth/register", handler.Register)
	e.POST("/auth/login", handler.Login)

	// Защищённые
	protected := e.Group("")
	protected.Use(middlewares.JWT(jwtSecret))

	// Users
	protected.GET("/auth/history", handler.HistoryOfRoutes)

	// Rooms
	protected.POST("/rooms", handler.CreateRoom)
	protected.GET("/rooms", handler.FindRoom)
	protected.GET("/rooms/:id", handler.GetRoomDetails)
	protected.POST("/rooms/:id/join", handler.JoinRoom)
	protected.POST("/rooms/:id/exit", handler.ExitRoom)
	protected.POST("/rooms/:id/complete", handler.CompleteRide) // триггер оплаты

	// Payments
	protected.POST("/payments/process", handler.ProcessPayment)
	protected.POST("/payments/refund", handler.RefundPayment)
	protected.GET("/payments/history", handler.GetPaymentHistory)
}
