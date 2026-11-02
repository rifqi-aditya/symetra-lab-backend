package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	AppEnv             string
	SupabaseURL        string
	SupabaseAnonKey    string
	ShopeePartnerID      int64
	ShopeePartnerKey     string
	ShopeePushPartnerKey string
	ShopeeIsProduction   bool
	ShopeeRedirectURL    string
	TikTokAppKey         string
	TikTokAppSecret      string
	TikTokIsProduction   bool
	TikTokRedirectURL    string
	FrontendURL          string
	AllowedOrigins       []string
	EncryptionKey        string
}

func Load() *Config {
	// Try loading .env from current directory or parent
	if err := godotenv.Load(".env"); err != nil {
		if err2 := godotenv.Load("../../.env"); err2 != nil {
			log.Printf("[Config] Notice: .env file not loaded (%v, %v), using environment variables", err, err2)
		} else {
			log.Println("[Config] Loaded .env from ../../.env")
		}
	} else {
		log.Println("[Config] Loaded .env from current directory")
	}

	partnerIDStr := getEnv("SHOPEE_PARTNER_ID", "0")
	partnerID, _ := strconv.ParseInt(partnerIDStr, 10, 64)
	isProdStr := getEnv("SHOPEE_IS_PRODUCTION", "false")
	isProd, _ := strconv.ParseBool(isProdStr)

	tikTokIsProdStr := getEnv("TIKTOK_IS_PRODUCTION", "false")
	tikTokIsProd, _ := strconv.ParseBool(tikTokIsProdStr)

	return &Config{
		Port:                 getEnv("PORT", "8081"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		AppEnv:               getEnv("APP_ENV", "development"),
		SupabaseURL:          getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:      getEnv("SUPABASE_ANON_KEY", ""),
		ShopeePartnerID:      partnerID,
		ShopeePartnerKey:     getEnv("SHOPEE_PARTNER_KEY", ""),
		ShopeePushPartnerKey: getEnv("SHOPEE_PUSH_PARTNER_KEY", ""),
		ShopeeIsProduction:   isProd,
		ShopeeRedirectURL:    getEnv("SHOPEE_REDIRECT_URL", "http://localhost:8080/api/v1/shopee/callback"),
		TikTokAppKey:         getEnv("TIKTOK_APP_KEY", ""),
		TikTokAppSecret:      getEnv("TIKTOK_APP_SECRET", ""),
		TikTokIsProduction:   tikTokIsProd,
		TikTokRedirectURL:    getEnv("TIKTOK_REDIRECT_URL", "http://localhost:8080/api/v1/tokopedia/callback"),
		FrontendURL:          getEnv("FRONTEND_URL", "http://localhost:5173"),
		AllowedOrigins:       parseOrigins(getEnv("ALLOWED_ORIGINS", "")),
		EncryptionKey:        getEnv("ENCRYPTION_KEY", "symetra-lab-default-encryption-secret-key-32"),
	}
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var result []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return strings.TrimSpace(val)
	}
	return strings.TrimSpace(defaultVal)
}
