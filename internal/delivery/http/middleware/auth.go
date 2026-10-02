package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/dto"
)

type supabaseUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// AuthMiddleware creates a middleware that verifies authentication via Supabase JWT or X-User-ID.
func AuthMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	httpClient := &http.Client{Timeout: 5 * time.Second}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if token != "" && cfg.SupabaseURL != "" {
					supabaseEndpoint := strings.TrimRight(cfg.SupabaseURL, "/") + "/auth/v1/user"
					req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, supabaseEndpoint, nil)
					if err == nil {
						req.Header.Set("apikey", cfg.SupabaseAnonKey)
						req.Header.Set("Authorization", "Bearer "+token)

						resp, err := httpClient.Do(req)
						if err == nil && resp.StatusCode == http.StatusOK {
							defer resp.Body.Close()
							var u supabaseUserResponse
							if err := json.NewDecoder(resp.Body).Decode(&u); err == nil && u.ID != "" {
								if parsedID, err := uuid.Parse(u.ID); err == nil {
									c.Set("user_id", parsedID)
									c.Set("user_email", u.Email)
									c.Set("user_role", u.Role)
									return next(c)
								}
							}
						}
					}
				}
			}

			// Fallback: X-User-ID header (developer testing or internal service call)
			if xUserID := c.Request().Header.Get("X-User-ID"); xUserID != "" {
				if parsedID, err := uuid.Parse(xUserID); err == nil {
					c.Set("user_id", parsedID)
					return next(c)
				}
			}

			return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized: valid token or credentials required", "Missing or invalid authorization token"))
		}
	}
}