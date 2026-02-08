package handler

import (
	"net/http"
	"time"

	"github.com/gigaonion/taskalyst/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TimeHandler struct {
	u usecase.TimeUsecase
}

func NewTimeHandler(u usecase.TimeUsecase) *TimeHandler {
	return &TimeHandler{u: u}
}

type StartTimerRequest struct {
	ProjectID string `json:"project_id" validate:"required"`
	TaskID    string `json:"task_id"`
	Note      string `json:"note"`
}

type DateRangeRequest struct {
	From string `query:"from"`
	To   string `query:"to"`
}

func (h *TimeHandler) StartTimer(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	var req StartTimerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	pID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	var tID *uuid.UUID
	if req.TaskID != "" {
		id, err := uuid.Parse(req.TaskID)
		if err == nil {
			tID = &id
		}
	}

	entry, err := h.u.StartTimeEntry(c.Request().Context(), userID, pID, tID, req.Note)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, entry)
}

func (h *TimeHandler) StopTimer(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid entry id")
	}

	entry, err := h.u.StopTimeEntry(c.Request().Context(), userID, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, entry)
}

func (h *TimeHandler) GetStats(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)

	to := time.Now()
	from := to.AddDate(-1, 0, 0)

	if f := c.QueryParam("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			from = t
		}
	}
	if t := c.QueryParam("to"); t != "" {
		if tm, err := time.Parse("2006-01-02", t); err == nil {
			to = tm
		}
	}

	stats, err := h.u.GetContributionStats(c.Request().Context(), userID, from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, stats)
}
