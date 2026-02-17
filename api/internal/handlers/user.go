package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"we_ride/api/internal/clients"
	"we_ride/api/internal/config"
	pb "we_ride/internal/services/user_service/protoc/gen/go"
)

type APIHandler struct {
	userService *clients.UserServiceClient
}

func NewAPIHandler(userService *clients.UserServiceClient) *APIHandler {
	return &APIHandler{userService: userService}
}

func (h *APIHandler) Register(c echo.Context) error {
	var req pb.RegisterRequest // Здесь используется правильный тип из сгенерированного кода
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
	resp, err := h.userService.HistoryOfRoutes(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get routes"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) InitRoutes(e *echo.Echo) {
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)
	e.GET("/routes", h.HistoryOfRoutes)
}

func NewServer(userService *clients.UserServiceClient, cfg *config.Config) *echo.Echo {
	e := echo.New()

	handler := NewAPIHandler(userService)
	handler.InitRoutes(e)

	return e
}
