package handler

import (
	"net/http"

	"github.com/gigaonion/taskalyst/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	u usecase.UserUsecase
}

func NewUserHandler(u usecase.UserUsecase) *UserHandler {
	return &UserHandler{u: u}
}

type SignUpRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// SignUp
func (h *UserHandler) SignUp(c echo.Context) error {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.u.SignUp(c.Request().Context(), req.Email, req.Password, req.Name)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(http.StatusCreated, UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
		Role:  string(user.Role),
	})
}

// Login
func (h *UserHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	tokenPair, err := h.u.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(http.StatusOK, tokenPair)
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID := getUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user id from context")
	}

	user, err := h.u.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
		Role:  string(user.Role),
	})
}
