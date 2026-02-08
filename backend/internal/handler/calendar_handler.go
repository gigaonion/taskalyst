package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/usecase"
)

type CalendarHandler struct {
	u usecase.CalendarUsecase
}

func NewCalendarHandler(u usecase.CalendarUsecase) *CalendarHandler {
	return &CalendarHandler{u: u}
}

type CreateCalendarRequest struct {
	Name        string `json:"name" validate:"required"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

type CreateEventRequest struct {
	ProjectID   string    `json:"project_id" validate:"required"`
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartAt     time.Time `json:"start_at" validate:"required"`
	EndAt       time.Time `json:"end_at" validate:"required"`
	IsAllDay    bool      `json:"is_all_day"`
}

type CreateTimetableSlotRequest struct {
	ProjectID string `json:"project_id" validate:"required"`
	DayOfWeek int    `json:"day_of_week" validate:"min=0,max=6"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time" validate:"required"`
	Location  string `json:"location"`
}

func (h *CalendarHandler) CreateEvent(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	var req CreateEventRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	pid, _ := uuid.Parse(req.ProjectID)
	event, err := h.u.CreateEvent(c.Request().Context(), userID, pid, req.Title, req.Description, req.Location, req.StartAt, req.EndAt, req.IsAllDay)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, event)
}
func (h *CalendarHandler) ListEvents(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	events, err := h.u.ListEvents(c.Request().Context(), userID, start, end)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, events)
}

func (h *CalendarHandler) CreateTimetableSlot(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	var req CreateTimetableSlotRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	pid, _ := uuid.Parse(req.ProjectID)
	
	layout := "15:04"
	start, _ := time.Parse(layout, req.StartTime)
	end, _ := time.Parse(layout, req.EndTime)

	slot, err := h.u.CreateTimetableSlot(c.Request().Context(), userID, pid, int32(req.DayOfWeek), start, end, req.Location)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, slot)
}

func (h *CalendarHandler) ListTimetable(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	slots, err := h.u.ListTimetable(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, slots)
}

func (h *CalendarHandler) SyncSchedule(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	dateStr := c.QueryParam("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	if err := h.u.SyncDailySchedule(c.Request().Context(), userID, date); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "synced"})
}

func (h *CalendarHandler) CreateCalendar(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	var req CreateCalendarRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	calendar, err := h.u.CreateCalendar(c.Request().Context(), userID, req.Name, req.Color, req.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, calendar)
}

func (h *CalendarHandler) ListCalendars(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	calendars, err := h.u.ListCalendars(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, calendars)
}

func (h *CalendarHandler) DeleteCalendar(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid calendar id")
	}

	if err := h.u.DeleteCalendar(c.Request().Context(), userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
