package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/config"
	"github.com/gigaonion/taskalyst/backend/internal/handler/middleware"
)

func RegisterRoutes(e *echo.Echo, userHandler *UserHandler, projectHandler *ProjectHandler, cfg *config.Config) {
	// Auth Group
	authGroup := e.Group("/auth")
	authGroup.POST("/signup", userHandler.SignUp)
	authGroup.POST("/login", userHandler.Login)

	api := e.Group("/api")
	api.Use(middleware.JWTMiddleware(cfg))

	api.GET("/users/me", userHandler.GetMe)
	api.POST("/categories", projectHandler.CreateCategory)
  api.GET("/categories", projectHandler.ListCategories)

  api.POST("/projects", projectHandler.CreateProject)
  api.GET("/projects", projectHandler.ListProjects)
}
