package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/pkg/auth"
)
type ApiTokenHandler struct {
	repo *repository.Queries // Usecase層を飛ばして簡易実装する場合
}

func NewApiTokenHandler(repo *repository.Queries) *ApiTokenHandler {
	return &ApiTokenHandler{repo: repo}
}

type CreateTokenRequest struct {
	Name string `json:"name" validate:"required"`
}

func (h *ApiTokenHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	var req CreateTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	// Generate
	rawToken, tokenHash, err := auth.GeneratePAT()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	// Save to DB
	_, err = h.repo.CreateApiToken(c.Request().Context(), repository.CreateApiTokenParams{
		UserID:    userID,
		Name:      req.Name,
		TokenHash: tokenHash,
	ExpiresAt: pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// ユーザーには生のトークンを返す
	return c.JSON(http.StatusCreated, map[string]string{
		"token": rawToken,
		"name":  req.Name,
	})
}

func (h *ApiTokenHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	tokens, err := h.repo.ListApiTokens(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, tokens)
}

func (h *ApiTokenHandler) Revoke(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.repo.DeleteApiToken(c.Request().Context(), repository.DeleteApiTokenParams{
		ID: id, UserID: userID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
