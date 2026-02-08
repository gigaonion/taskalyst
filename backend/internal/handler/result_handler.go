package handler

import (
	"net/http"
	"time"

	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type ResultHandler struct {
	u usecase.ResultUsecase
}

func NewResultHandler(u usecase.ResultUsecase) *ResultHandler {
	return &ResultHandler{u: u}
}

// リクエストボディ
type CreateResultRequest struct {
	ProjectID  string  `json:"project_id" validate:"required"`
	TaskID     string  `json:"task_id"`
	Type       string  `json:"type" validate:"required"`
	Value      float64 `json:"value" validate:"required"`
	RecordedAt string  `json:"recorded_at"`
	Note       string  `json:"note"`
}

// レスポンス(pgtype.Numericをfloat64に変換)
type ResultResponse struct {
	ID           string  `json:"id"`
	ProjectID    string  `json:"project_id"`
	TaskID       string  `json:"task_id,omitempty"`
	Type         string  `json:"type"`
	Value        float64 `json:"value"`
	RecordedAt   string  `json:"recorded_at"`
	Note         string  `json:"note"`
	ProjectTitle string  `json:"project_title,omitempty"`
	ProjectColor string  `json:"project_color,omitempty"`
	TaskTitle    string  `json:"task_title,omitempty"`
}

func toFloat64(n pgtype.Numeric) float64 {
	f, _ := n.Float64Value()
	return f.Float64
}

func toResultResponse(r *repository.Result) ResultResponse {
	taskID := ""
	if r.TargetTaskID.Valid {
		taskID = uuid.UUID(r.TargetTaskID.Bytes).String()
	}

	return ResultResponse{
		ID:         r.ID.String(),
		ProjectID:  r.ProjectID.String(),
		TaskID:     taskID,
		Type:       r.Type,
		Value:      toFloat64(r.Value),
		RecordedAt: r.RecordedAt.Time.Format(time.RFC3339),
		Note:       r.Note.String,
	}
}

func toResultListResponse(r repository.ListResultsRow) ResultResponse {
	taskID := ""
	if r.TargetTaskID.Valid {
		taskID = uuid.UUID(r.TargetTaskID.Bytes).String()
	}

	return ResultResponse{
		ID:           r.ID.String(),
		ProjectID:    r.ProjectID.String(),
		TaskID:       taskID,
		Type:         r.Type,
		Value:        toFloat64(r.Value),
		RecordedAt:   r.RecordedAt.Time.Format(time.RFC3339),
		Note:         r.Note.String,
		ProjectTitle: r.ProjectTitle,
		ProjectColor: r.ProjectColor.String,
		TaskTitle:    r.TaskTitle.String,
	}
}

func (h *ResultHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)

	var req CreateResultRequest
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

	recordedAt := time.Now()
	if req.RecordedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.RecordedAt); err == nil {
			recordedAt = t
		}
	}

	result, err := h.u.CreateResult(c.Request().Context(), userID, pID, tID, req.Type, req.Value, recordedAt, req.Note)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, toResultResponse(result))
}

func (h *ResultHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)

	to := time.Now()
	from := to.AddDate(0, -1, 0)

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

	var pID *uuid.UUID
	if pid := c.QueryParam("project_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err == nil {
			pID = &id
		}
	}

	rows, err := h.u.ListResults(c.Request().Context(), userID, pID, from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := make([]ResultResponse, len(rows))
	for i, row := range rows {
		res[i] = toResultListResponse(row)
	}

	return c.JSON(http.StatusOK, res)
}

func (h *ResultHandler) Delete(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.u.DeleteResult(c.Request().Context(), userID, id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
