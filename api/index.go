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

	// Dapatkan path asli yang diminta
	requestedPath := ""
	if qPath := r.URL.Query().Get("path"); qPath != "" {
		requestedPath = qPath
	} else if origPath := r.Header.Get("x-matched-path"); origPath != "" {
		requestedPath = origPath
	} else if origPath := r.Header.Get("x-vercel-matched-path"); origPath != "" {
		requestedPath = origPath
	}

	if requestedPath != "" {
		// Bersihkan query param 'path' dari URL query jika ada
		q := r.URL.Query()
		q.Del("path")
		newRawQuery := q.Encode()

		r.URL.Path = requestedPath
		r.URL.RawQuery = newRawQuery
		if newRawQuery != "" {
			r.RequestURI = requestedPath + "?" + newRawQuery
		} else {
			r.RequestURI = requestedPath
		}
	}

	engine.ServeHTTP(w, r)
}
