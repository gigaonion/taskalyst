package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/config"
	"github.com/gigaonion/taskalyst/backend/internal/infra/repository"
	"github.com/gigaonion/taskalyst/backend/pkg/auth"
)

func AuthMiddleware(cfg *config.Config, repo *repository.Queries) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// PAT
			apiKey := c.Request().Header.Get("X-API-KEY")
			if apiKey != "" {
				hash := auth.HashPAT(apiKey)
				user, err := repo.GetUserByTokenHash(c.Request().Context(), hash)
				if err == nil {
					c.Set("user_id", user.ID)
					c.Set("role", string(user.Role))
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
			}

			// JWT
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					claims, err := auth.ValidateToken(parts[1], cfg.JWTSecret)
					if err == nil {
						c.Set("user_id", claims.UserID)
						c.Set("role", claims.Role)
						return next(c)
					}
				}
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid credentials")
		}
	}
}
