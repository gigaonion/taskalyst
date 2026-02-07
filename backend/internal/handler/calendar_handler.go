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

type CreateEventRequest struct {
	ProjectID string    `json:"project_id" validate:"required"`
	Title     string    `json:"title" validate:"required"`
	StartAt   time.Time `json:"start_at" validate:"required"`
	EndAt     time.Time `json:"end_at" validate:"required"`
	IsAllDay  bool      `json:"is_all_day"`
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
	event, err := h.u.CreateEvent(c.Request().Context(), userID, pid, req.Title, req.StartAt, req.EndAt, req.IsAllDay)
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
