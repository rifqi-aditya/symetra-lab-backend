package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/middleware"
)

func TestAuthMiddleware_XUserIDGating(t *testing.T) {
	testUserID := uuid.New().String()

	t.Run("should reject X-User-ID when APP_ENV is production", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("X-User-ID", testUserID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		cfg := &config.Config{
			AppEnv: "production",
		}

		mw := middleware.AuthMiddleware(cfg)
		handlerCalled := false
		handler := mw(func(c echo.Context) error {
			handlerCalled = true
			return c.NoContent(http.StatusOK)
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if handlerCalled {
			t.Fatalf("Handler must NOT be called in production with X-User-ID")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("should allow X-User-ID when APP_ENV is development", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("X-User-ID", testUserID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		cfg := &config.Config{
			AppEnv: "development",
		}

		mw := middleware.AuthMiddleware(cfg)
		handlerCalled := false
		handler := mw(func(c echo.Context) error {
			handlerCalled = true
			uidVal := c.Get("user_id")
			if uidVal == nil {
				t.Fatalf("expected user_id in context, got nil")
			}
			if uidVal.(uuid.UUID).String() != testUserID {
				t.Fatalf("expected user_id %s, got %s", testUserID, uidVal.(uuid.UUID).String())
			}
			return c.NoContent(http.StatusOK)
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !handlerCalled {
			t.Fatalf("Handler SHOULD be called in development with valid X-User-ID")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
