package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/internal/usecase"
)

type ProjectHandler struct {
	u usecase.ProjectUsecase
}

func NewProjectHandler(u usecase.ProjectUsecase) *ProjectHandler {
	return &ProjectHandler{u: u}
}


type CreateCategoryRequest struct {
	Name     string `json:"name" validate:"required"`
	RootType string `json:"root_type" validate:"required,oneof=GROWTH LIFE WORK HOBBY OTHER"`
	Color    string `json:"color"`
}

type CreateProjectRequest struct {
	CategoryID string `json:"category_id" validate:"required"`
	Title      string `json:"title" validate:"required"`
	Description string `json:"description"`
}


func (h *ProjectHandler) CreateCategory(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	
	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	// バリデーション
	if err := c.Validate(&req); err != nil {
		return err
	}
	rootType := repository.RootCategoryType(req.RootType)

	category, err := h.u.CreateCategory(c.Request().Context(), userID, req.Name, rootType, req.Color)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, category)
}

func (h *ProjectHandler) ListCategories(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)

	categories, err := h.u.ListCategories(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, categories)
}

func (h *ProjectHandler) CreateProject(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)

	var req CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid category id")
	}

	project, err := h.u.CreateProject(c.Request().Context(), userID, catID, req.Title, req.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) ListProjects(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	
	// ?archived=true
	var isArchived *bool
	if c.QueryParam("archived") == "true" {
		t := true
		isArchived = &t
	} else if c.QueryParam("archived") == "false" {
		f := false
		isArchived = &f
	}

	projects, err := h.u.ListProjects(c.Request().Context(), userID, isArchived)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, projects)
}
