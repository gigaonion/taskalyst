package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/config"
	"github.com/gigaonion/taskalyst/backend/internal/handler/middleware"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
)

func RegisterRoutes(e *echo.Echo, userHandler *UserHandler, projectHandler *ProjectHandler, taskHandler *TaskHandler, timeHandler *TimeHandler, apiTokenHandler *ApiTokenHandler, cfg *config.Config, calendarHandler *CalendarHandler,resultHandler *ResultHandler, repo *repository.Queries) {
	// Auth Group
	authGroup := e.Group("/auth")
	authGroup.POST("/signup", userHandler.SignUp)
	authGroup.POST("/login", userHandler.Login)

	api := e.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg, repo))

	api.GET("/users/me", userHandler.GetMe)
	
	// API Tokens
	api.POST("/tokens", apiTokenHandler.Create)
	api.GET("/tokens", apiTokenHandler.List)
	api.DELETE("/tokens/:id", apiTokenHandler.Revoke)

	api.POST("/categories", projectHandler.CreateCategory)
	api.GET("/categories", projectHandler.ListCategories)

	api.POST("/projects", projectHandler.CreateProject)
	api.GET("/projects", projectHandler.ListProjects)

	api.POST("/tasks", taskHandler.CreateTask)
	api.GET("/tasks", taskHandler.ListTasks)
	api.PATCH("/tasks/:id/status", taskHandler.UpdateTaskStatus)

	api.POST("/tasks/:id/checklist", taskHandler.AddChecklistItem)
	api.PATCH("/checklist-items/:id", taskHandler.ToggleChecklistItem)

	api.POST("/time-entries", timeHandler.StartTimer)
	api.PATCH("/time-entries/:id/stop", timeHandler.StopTimer)
	api.GET("/stats/growth", timeHandler.GetStats)

	api.POST("/events", calendarHandler.CreateEvent)
	api.GET("/events", calendarHandler.ListEvents)

	api.POST("/timetable", calendarHandler.CreateTimetableSlot)
	api.GET("/timetable", calendarHandler.ListTimetable)

	api.POST("/api-tokens", apiTokenHandler.Create)
	api.GET("/api-tokens", apiTokenHandler.List)
	api.DELETE("/api-tokens/:id", apiTokenHandler.Revoke)

	api.POST("/sync/schedule", calendarHandler.SyncSchedule)

	api.POST("/results", resultHandler.Create)
  api.GET("/results", resultHandler.List)
  api.DELETE("/results/:id", resultHandler.Delete)
	
}
