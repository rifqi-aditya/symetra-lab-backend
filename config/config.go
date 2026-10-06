package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	AppEnv             string
	SupabaseURL        string
	SupabaseAnonKey    string
	ShopeePartnerID    int64
	ShopeePartnerKey   string
	ShopeeIsProduction bool
	ShopeeRedirectURL  string
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

	return &Config{
		Port:               getEnv("PORT", "8081"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		AppEnv:             getEnv("APP_ENV", "development"),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
		ShopeePartnerID:    partnerID,
		ShopeePartnerKey:   getEnv("SHOPEE_PARTNER_KEY", ""),
		ShopeeIsProduction: isProd,
		ShopeeRedirectURL:  getEnv("SHOPEE_REDIRECT_URL", "http://localhost:8080/api/v1/shopee/callback"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
