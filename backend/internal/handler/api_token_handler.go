package handler

import (
	"net/http"

	"github.com/gigaonion/taskalyst/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ApiTokenHandler struct {
	u usecase.ApiTokenUsecase
}

func NewApiTokenHandler(u usecase.ApiTokenUsecase) *ApiTokenHandler {
	return &ApiTokenHandler{u: u}
}

type CreateTokenRequest struct {
	Name string `json:"name" validate:"required"`
}

func (h *ApiTokenHandler) Create(c echo.Context) error {
	userID := getUserID(c)
	var req CreateTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	rawToken, _, err := h.u.Create(c.Request().Context(), userID, req.Name)
	if err != nil {
		return HandleError(c, err)
	}

	// ユーザーには生のトークンを返す
	return c.JSON(http.StatusCreated, map[string]string{
		"token": rawToken,
		"name":  req.Name,
	})
}

func (h *ApiTokenHandler) List(c echo.Context) error {
	userID := getUserID(c)
	tokens, err := h.u.List(c.Request().Context(), userID)
	if err != nil {
		return HandleError(c, err)
	}
	return c.JSON(http.StatusOK, tokens)
}

func (h *ApiTokenHandler) Revoke(c echo.Context) error {
	userID := getUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.u.Revoke(c.Request().Context(), userID, id); err != nil {
		return HandleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
