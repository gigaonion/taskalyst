package handler

import (
	"errors"
	"net/http"

	"github.com/gigaonion/taskalyst/backend/internal/usecase"
	"github.com/labstack/echo/v4"
)

func HandleError(c echo.Context, err error) error {
	if err == nil {
		return nil
	}

	var domainErr *usecase.DomainError
	if errors.As(err, &domainErr) {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(domainErr.Err, usecase.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(domainErr.Err, usecase.ErrUnauthorized):
			status = http.StatusUnauthorized
		case errors.Is(domainErr.Err, usecase.ErrForbidden):
			status = http.StatusForbidden
		case errors.Is(domainErr.Err, usecase.ErrConflict):
			status = http.StatusConflict
		case errors.Is(domainErr.Err, usecase.ErrBadRequest):
			status = http.StatusBadRequest
		}
		return echo.NewHTTPError(status, domainErr.Message)
	}

	// Default to internal server error
	return HandleError(c, err)
}
