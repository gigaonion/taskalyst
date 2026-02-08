package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// parseDateQuery parses a date string from query parameters.
func parseDateQuery(c echo.Context, name string, defaultValue time.Time) time.Time {
	val := c.QueryParam(name)
	if val == "" {
		return defaultValue
	}
	t, err := time.Parse("2006-01-02", val)
	if err != nil {
		return defaultValue
	}
	return t
}

// getUserID extracts the user_id from the echo context.
func getUserID(c echo.Context) uuid.UUID {
	id, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		// This should not happen if AuthMiddleware is used
		return uuid.Nil
	}
	return id
}
