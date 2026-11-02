package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/middleware"
)

func TestSetupCORS(t *testing.T) {
	t.Run("should allow configured frontend URL in production", func(t *testing.T) {
		cfg := &config.Config{
			AppEnv:      "production",
			FrontendURL: "https://symetralab-app.vercel.app",
		}

		e := echo.New()
		e.Use(middleware.SetupCORS(cfg))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://symetralab-app.vercel.app")
		req.Header.Set("Access-Control-Request-Method", "GET")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "https://symetralab-app.vercel.app" {
			t.Fatalf("expected origin allowed, got: %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("should reject unknown malicious origin in production", func(t *testing.T) {
		cfg := &config.Config{
			AppEnv:      "production",
			FrontendURL: "https://symetralab-app.vercel.app",
		}

		e := echo.New()
		e.Use(middleware.SetupCORS(cfg))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://malicious-site.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("malicious origin must NOT be allowed, got: %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("should allow localhost in development", func(t *testing.T) {
		cfg := &config.Config{
			AppEnv: "development",
		}

		e := echo.New()
		e.Use(middleware.SetupCORS(cfg))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", "GET")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
			t.Fatalf("expected localhost allowed in dev, got: %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("should allow symetra-lab-workspace.vercel.app in production", func(t *testing.T) {
		cfg := &config.Config{
			AppEnv: "production",
		}

		e := echo.New()
		e.Use(middleware.SetupCORS(cfg))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		origins := []string{
			"https://symetra-lab-workspace.vercel.app",
			"https://symetra-lab-workspace-git-main.vercel.app",
			"https://symetralab.vercel.app",
		}

		for _, orig := range origins {
			req := httptest.NewRequest(http.MethodOptions, "/test", nil)
			req.Header.Set("Origin", orig)
			req.Header.Set("Access-Control-Request-Method", "POST")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Header().Get("Access-Control-Allow-Origin") != orig {
				t.Fatalf("expected %s allowed in production, got: %s", orig, rec.Header().Get("Access-Control-Allow-Origin"))
			}
		}
	})
}

