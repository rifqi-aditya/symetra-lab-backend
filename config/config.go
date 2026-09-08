package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port               string
	AppEnv             string
	DatabaseURL        string
	ShopeePartnerID    int64
	ShopeePartnerKey   string
	ShopeeIsProduction bool
	ShopeeRedirectURL  string
}

func LoadConfig() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] File .env tidak ditemukan, membaca dari environment sistem")
	}
	// 2. Baca string angka partner_id dan ubah ke int64
	partnerIDStr := getEnv("SHOPEE_PARTNER_ID", "0")
	partnerID, _ := strconv.ParseInt(partnerIDStr, 10, 64)
	// 3. Baca boolean apakah mode produksi atau sandbox
	isProdStr := getEnv("SHOPEE_IS_PRODUCTION", "false")
	isProd, _ := strconv.ParseBool(isProdStr)
	return &Config{
		Port:               getEnv("PORT", "8080"),
		AppEnv:             getEnv("APP_ENV", "development"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		ShopeePartnerID:    partnerID,
		ShopeePartnerKey:   getEnv("SHOPEE_PARTNER_KEY", ""),
		ShopeeIsProduction: isProd,
		ShopeeRedirectURL:  getEnv("SHOPEE_REDIRECT_URL", "http://localhost:8080/api/v1/shopee/callback"),
	}
}

// getEnv adalah fungsi pembantu untuk membaca env, jika kosong pakai nilai bawaan (fallback)
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
