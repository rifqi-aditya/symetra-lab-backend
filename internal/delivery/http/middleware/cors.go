package middleware

import (
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"symetra-lab-backend-v2/config"
)

func SetupCORS(cfg *config.Config) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			if origin == "" {
				return false, nil
			}

			// 1. Check configured frontend URL
			if cfg != nil && cfg.FrontendURL != "" {
				if strings.TrimRight(origin, "/") == strings.TrimRight(cfg.FrontendURL, "/") {
					return true, nil
				}
			}

			// 2. Check explicitly configured allowed origins
			if cfg != nil {
				for _, allowed := range cfg.AllowedOrigins {
					if strings.TrimRight(origin, "/") == strings.TrimRight(allowed, "/") {
						return true, nil
					}
				}
			}

			// 3. Parse origin URL
			parsed, err := url.Parse(origin)
			if err != nil {
				return false, nil
			}
			hostname := parsed.Hostname()

			// 4. In development, allow localhost and testing tunnels
			isProd := cfg != nil && (strings.EqualFold(cfg.AppEnv, "production") || strings.EqualFold(cfg.AppEnv, "prod"))
			if !isProd {
				if hostname == "localhost" || hostname == "127.0.0.1" || strings.HasSuffix(hostname, ".ngrok-free.app") {
					return true, nil
				}
			}

			// 5. Whitelist trusted Vercel frontend domains for symetra-lab
			if strings.HasSuffix(hostname, ".vercel.app") {
				if strings.HasPrefix(hostname, "symetra-lab") || strings.HasPrefix(hostname, "symetralab") {
					return true, nil
				}
			}

			return false, nil
		},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-User-ID", "x-user-id", "X-Requested-With", "Cache-Control"},
	})
}
