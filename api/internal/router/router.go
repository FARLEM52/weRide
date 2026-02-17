package router

import (
	"github.com/labstack/echo/v4"
	"we_ride/api/internal/clients"
	"we_ride/api/internal/handlers"
	"we_ride/api/internal/middlewares"
)

func InitRoutes(e *echo.Echo, userService *clients.UserServiceClient, jwtSecret []byte) {

	handler := handlers.NewAPIHandler(userService)

	e.POST("/auth/register", handler.Register)
	e.POST("/auth/login", handler.Login)
	authGroup := e.Group("/auth", middlewares.JWT(jwtSecret))
	authGroup.GET("/history", handler.HistoryOfRoutes)
}
