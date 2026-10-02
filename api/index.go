package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/app"
)

var (
	echoEngine *echo.Echo
	initOnce   sync.Once
	initErr    error
)

func getEchoEngine() (*echo.Echo, error) {
	initOnce.Do(func() {
		echoEngine, _, initErr = app.InitEchoApp()
		if initErr != nil {
			log.Printf("[Vercel Handler] Error initializing Echo app: %v", initErr)
			return
		}

		// Daftarkan Pre-middleware untuk rewrite path URL dari query param atau header Vercel
		echoEngine.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				req := c.Request()
				requestedPath := ""

				if qPath := req.URL.Query().Get("path"); qPath != "" {
					requestedPath = qPath
				} else if origPath := req.Header.Get("x-matched-path"); origPath != "" {
					requestedPath = origPath
				} else if origPath := req.Header.Get("x-vercel-matched-path"); origPath != "" {
					requestedPath = origPath
				}

				if requestedPath != "" {
					q := req.URL.Query()
					q.Del("path")
					req.URL.Path = requestedPath
					req.URL.RawPath = requestedPath
					req.URL.RawQuery = q.Encode()
					if q.Encode() != "" {
						req.RequestURI = requestedPath + "?" + q.Encode()
					} else {
						req.RequestURI = requestedPath
					}
				}
				return next(c)
			}
		})
	})
	return echoEngine, initErr
}

// Handler adalah entry point standar Serverless Function untuk Vercel Go runtime
func Handler(w http.ResponseWriter, r *http.Request) {
	// Dapatkan path asli yang diminta
	requestedPath := ""
	if qPath := r.URL.Query().Get("path"); qPath != "" {
		requestedPath = qPath
	} else if origPath := r.Header.Get("x-matched-path"); origPath != "" {
		requestedPath = origPath
	} else if origPath := r.Header.Get("x-vercel-matched-path"); origPath != "" {
		requestedPath = origPath
	}

	// Respon cepat untuk /health check
	if requestedPath == "/health" || r.URL.Path == "/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","version":"2.0.0","runtime":"vercel"}`))
		return
	}

	engine, err := getEchoEngine()
	if err != nil {
		http.Error(w, "Internal Server Error: Failed to initialize application ("+err.Error()+")", http.StatusInternalServerError)
		return
	}

	if requestedPath != "" {
		q := r.URL.Query()
		q.Del("path")
		r.URL.Path = requestedPath
		r.URL.RawPath = requestedPath
		r.URL.RawQuery = q.Encode()
		if q.Encode() != "" {
			r.RequestURI = requestedPath + "?" + q.Encode()
		} else {
			r.RequestURI = requestedPath
		}
	}

	engine.ServeHTTP(w, r)
}
