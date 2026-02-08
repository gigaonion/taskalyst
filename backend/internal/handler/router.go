package handler

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/config"
	"github.com/gigaonion/taskalyst/backend/internal/handler/middleware"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
)

func RegisterRoutes(e *echo.Echo, userHandler *UserHandler, projectHandler *ProjectHandler, taskHandler *TaskHandler, timeHandler *TimeHandler, apiTokenHandler *ApiTokenHandler, cfg *config.Config, calendarHandler *CalendarHandler,resultHandler *ResultHandler, caldavHandler *CalDavHandler, repo *repository.Queries) {
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

	api.POST("/calendars", calendarHandler.CreateCalendar)
	api.GET("/calendars", calendarHandler.ListCalendars)
	api.DELETE("/calendars/:id", calendarHandler.DeleteCalendar)

	api.POST("/timetable", calendarHandler.CreateTimetableSlot)
	api.GET("/timetable", calendarHandler.ListTimetable)

	api.POST("/sync/schedule", calendarHandler.SyncSchedule)

	api.POST("/results", resultHandler.Create)
	api.GET("/results", resultHandler.List)
	api.DELETE("/results/:id", resultHandler.Delete)

	// CalDAV
	dav := e.Group("/dav")
	dav.Use(middleware.AuthMiddleware(cfg, repo))
	dav.Match([]string{"OPTIONS"}, "", caldavHandler.Options)
	dav.Match([]string{"PROPFIND"}, "", caldavHandler.Propfind)
	dav.Match([]string{"PROPFIND"}, "/:calendarID", caldavHandler.Propfind)
	dav.GET("/:calendarID", caldavHandler.GetCalendar)
	dav.GET("/:calendarID/:resource", caldavHandler.GetResource)
	dav.Match([]string{"PUT"}, "/:calendarID/:resource", caldavHandler.PutResource)
	dav.Match([]string{"OPTIONS"}, "/:calendarID/:resource", caldavHandler.Options)
	dav.Match([]string{"PROPFIND"}, "/:calendarID/:resource", caldavHandler.Propfind)

	// Well-known
	e.GET("/.well-known/caldav", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/dav")
	})
	
}
