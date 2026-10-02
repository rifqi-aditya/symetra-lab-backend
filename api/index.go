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
		}
	})
	return echoEngine, initErr
}

// Handler adalah entry point standar Serverless Function untuk Vercel Go runtime
func Handler(w http.ResponseWriter, r *http.Request) {
	engine, err := getEchoEngine()
	if err != nil {
		http.Error(w, "Internal Server Error: Failed to initialize application ("+err.Error()+")", http.StatusInternalServerError)
		return
	}

	// Saat di-rewrite oleh Vercel dari /health -> /api/index, r.URL.Path menjadi "/api/index"
	// Vercel menyimpan path asli user di header "x-matched-path" atau "x-vercel-matched-path"
	if origPath := r.Header.Get("x-matched-path"); origPath != "" {
		r.URL.Path = origPath
		r.RequestURI = origPath
	} else if origPath := r.Header.Get("x-vercel-matched-path"); origPath != "" {
		r.URL.Path = origPath
		r.RequestURI = origPath
	}

	engine.ServeHTTP(w, r)
}
