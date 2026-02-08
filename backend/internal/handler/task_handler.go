package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/internal/usecase"
)

type TaskHandler struct {
	u usecase.TaskUsecase
}

func NewTaskHandler(u usecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{u: u}
}


type CreateTaskRequest struct {
	ProjectID string    `json:"project_id" validate:"required"`
	Title     string    `json:"title" validate:"required"`
	Note      string    `json:"note"`
	DueDate   *time.Time `json:"due_date"`
}

type UpdateTaskStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=TODO DOING DONE"`
}

type AddChecklistItemRequest struct {
	Content string `json:"content" validate:"required"`
}

type ToggleChecklistItemRequest struct {
	IsCompleted bool `json:"is_completed"`
}


func (h *TaskHandler) CreateTask(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)

	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	task, err := h.u.CreateTask(c.Request().Context(), userID, projectID, req.Title, req.Note, req.DueDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) ListTasks(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)

	var projectID *uuid.UUID
	if pid := c.QueryParam("project_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err == nil {
			projectID = &id
		}
	}

	var status *repository.TaskStatus
	if s := c.QueryParam("status"); s != "" {
		st := repository.TaskStatus(s)
		status = &st
	}

	var from, to *time.Time
	if f := c.QueryParam("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			from = &t
		}
	}
	if t := c.QueryParam("to"); t != "" {
		if tm, err := time.Parse("2006-01-02", t); err == nil {
			to = &tm
		}
	}
	tasks, err := h.u.ListTasks(c.Request().Context(), userID, projectID, status, from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTaskStatus(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	var req UpdateTaskStatusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	task, err := h.u.UpdateTaskStatus(c.Request().Context(), userID, taskID, repository.TaskStatus(req.Status))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, task)
}

//チェックリスト
func (h *TaskHandler) AddChecklistItem(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	var req AddChecklistItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	item, err := h.u.AddChecklistItem(c.Request().Context(), taskID, req.Content)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, item)
}

func (h *TaskHandler) ToggleChecklistItem(c echo.Context) error {
	itemID, err := uuid.Parse(c.Param("id")) // URL: /checklist-items/:id
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid item id")
	}

	var req ToggleChecklistItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	item, err := h.u.ToggleChecklistItem(c.Request().Context(), itemID, req.IsCompleted)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, item)
}
