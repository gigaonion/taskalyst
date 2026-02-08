package handler

import (
	"github.com/gigaonion/taskalyst/backend/internal/config"
	"github.com/gigaonion/taskalyst/backend/internal/handler/middleware"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/labstack/echo/v4"
	"net/http"
)

func RegisterRoutes(e *echo.Echo, userHandler *UserHandler, projectHandler *ProjectHandler, taskHandler *TaskHandler, timeHandler *TimeHandler, apiTokenHandler *ApiTokenHandler, cfg *config.Config, calendarHandler *CalendarHandler, resultHandler *ResultHandler, caldavHandler *CalDavHandler, repo *repository.Queries) {
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

	// Discovery and Principal
	dav.Match([]string{"OPTIONS", "PROPFIND"}, "/principals/:userID", caldavHandler.Principal)
	dav.Match([]string{"OPTIONS", "PROPFIND"}, "/principals/:userID/", caldavHandler.Principal)
	dav.Match([]string{"OPTIONS", "PROPFIND"}, "/calendars/:userID", caldavHandler.CalendarHome)
	dav.Match([]string{"OPTIONS", "PROPFIND"}, "/calendars/:userID/", caldavHandler.CalendarHome)

	// Calendar Collection
	dav.Match([]string{"OPTIONS", "PROPFIND", "REPORT"}, "/calendars/:userID/:calendarID", caldavHandler.CalendarCollection)
	dav.Match([]string{"OPTIONS", "PROPFIND", "REPORT"}, "/calendars/:userID/:calendarID/", caldavHandler.CalendarCollection)
	dav.GET("/calendars/:userID/:calendarID", caldavHandler.GetCalendar) // For manual download

	// Calendar Resource
	dav.Match([]string{"OPTIONS", "PROPFIND", "GET", "PUT", "DELETE"}, "/calendars/:userID/:calendarID/:resource", caldavHandler.CalendarResource)

	// Well-known
	e.GET("/.well-known/caldav", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/dav/principals/")
	})
	dav.Match([]string{"OPTIONS", "PROPFIND"}, "/principals/", caldavHandler.PrincipalDiscovery)
	dav.Match([]string{"OPTIONS", "PROPFIND"}, "/principals", caldavHandler.PrincipalDiscovery)

}
