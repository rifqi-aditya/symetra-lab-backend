package main

import (
	"log"

	"symetra-lab-backend-v2/internal/app"
)

func main() {
	e, cfg, err := app.InitEchoApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	log.Printf("[Symetra Lab v2] Server listening on http://localhost:%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
